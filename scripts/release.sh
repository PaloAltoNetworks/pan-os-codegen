#!/bin/bash
# scripts/release.sh - Automates the PAN-OS release process
#
# This script automates the process of releasing pango and terraform-provider-panos
# by running codegen, copying generated code, and creating version tags.
#
# Usage:
#   ./scripts/release.sh [--auto|--manual|--dry-run]
#
# Modes:
#   --auto      Fully automated mode (runs everything, creates tags, pushes to remote)
#   --manual    Interactive mode (prompts for confirmation at key steps, no auto-push)
#   --dry-run   Simulation mode (shows what would be done without making changes)

set -e  # Exit on error
set -o pipefail

# Configuration
MODE="manual"  # Default mode
CODEGEN_DIR="$HOME/workspace/pan-os-codegen"
PANGO_DIR="$HOME/workspace/pango"
TERRAFORM_DIR="$HOME/workspace/terraform-provider-panos"
GOFIX_SCRIPT="$HOME/workspace/zsh-scripts/go-mod-replace.sh"
FIXDOCS_SCRIPT="$HOME/workspace/zsh-scripts/fix-docs.go"
FIXDOCS_CONFIG="$HOME/workspace/zsh-scripts/config.json"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}$(date '+%Y-%m-%d %H:%M:%S')${NC} - $1"
}

log_success() {
    echo -e "${GREEN}$(date '+%Y-%m-%d %H:%M:%S')${NC} - $1"
}

log_warning() {
    echo -e "${YELLOW}$(date '+%Y-%m-%d %H:%M:%S')${NC} - $1"
}

log_error() {
    echo -e "${RED}$(date '+%Y-%m-%d %H:%M:%S')${NC} - $1"
}

# Parse command-line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --auto)
                MODE="auto"
                shift
                ;;
            --manual)
                MODE="manual"
                shift
                ;;
            --dry-run)
                MODE="dry-run"
                shift
                ;;
            -h|--help)
                echo "Usage: $0 [--auto|--manual|--dry-run]"
                echo ""
                echo "Modes:"
                echo "  --auto      Fully automated mode (runs everything, creates tags, pushes to remote)"
                echo "  --manual    Interactive mode (prompts for confirmation at key steps, no auto-push)"
                echo "  --dry-run   Simulation mode (shows what would be done without making changes)"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                echo "Use --help for usage information"
                exit 1
                ;;
        esac
    done
}

# Execute command based on mode
execute() {
    local cmd="$1"
    local description="$2"

    if [ "$MODE" = "dry-run" ]; then
        log_info "[DRY-RUN] $description"
        log_info "[DRY-RUN] Would execute: $cmd"
    else
        log_info "$description"
        eval "$cmd"
    fi
}

# Check if repository is clean
check_repo_clean() {
    local repo_dir="$1"
    local repo_name=$(basename "$repo_dir")

    cd "$repo_dir"

    if [ -n "$(git status --porcelain)" ]; then
        log_warning "Repository $repo_name has uncommitted changes"

        if [ "$MODE" = "auto" ]; then
            log_error "Auto mode requires clean repositories. Exiting."
            exit 1
        elif [ "$MODE" = "manual" ]; then
            echo -n "Stash changes in $repo_name? (y/n): "
            read -r response
            if [[ "$response" =~ ^[Yy]$ ]]; then
                execute "git stash" "Stashing changes in $repo_name"
            else
                log_error "Please commit or stash changes before continuing"
                exit 1
            fi
        fi
    else
        log_success "Repository $repo_name is clean"
    fi
}

# Check if npx is available
check_npx() {
    if ! command -v npx &> /dev/null; then
        log_error "npx not found. Please install Node.js to use conventional-changelog tooling."
        log_error "Alternatively, install with: brew install node"
        exit 1
    fi
}

# Determine next version using conventional commits with standard-version
determine_next_version() {
    local repo_dir="$1"
    cd "$repo_dir"

    # Get the last tag
    local last_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
    log_info "Last tag: $last_tag"

    # Get commits since last tag
    local commits=$(git log --oneline --no-merges "$last_tag..HEAD" --format="%s" 2>/dev/null || echo "")

    if [ -z "$commits" ]; then
        log_warning "No new commits since $last_tag"
        echo "$last_tag"
        return
    fi

    # Use npx standard-version to determine next version (dry-run mode)
    log_info "Using standard-version to determine next version..."
    local next_version=$(npx --yes standard-version --dry-run --silent 2>&1 | grep "tagging release" | sed -n 's/.*tagging release \(v[0-9.]*\)/\1/p' | head -1)

    # If standard-version didn't work, fall back to manual detection
    if [ -z "$next_version" ]; then
        log_warning "standard-version failed, using fallback version detection"

        # Parse current version
        local version="${last_tag#v}"
        IFS='.' read -r major minor patch <<< "$version"

        # Check for breaking changes (major bump)
        if echo "$commits" | grep -qE "^[^:]+!:|BREAKING CHANGE:"; then
            major=$((major + 1))
            minor=0
            patch=0
        # Check for features (minor bump)
        elif echo "$commits" | grep -qE "^feat:"; then
            minor=$((minor + 1))
            patch=0
        # Default to patch bump
        else
            patch=$((patch + 1))
        fi

        next_version="v${major}.${minor}.${patch}"
    fi

    log_info "Next version: $next_version"
    echo "$next_version"
}

# Run codegen
run_codegen() {
    log_info "Running codegen..."
    cd "$CODEGEN_DIR"

    execute "make clean" "Cleaning previous build"
    execute "make codegen" "Generating code"

    log_success "Codegen completed"
}

# Copy generated code
copy_generated_code() {
    local source_dir="$1"
    local dest_dir="$2"
    local name="$3"

    log_info "Copying generated code to $name..."

    if [ "$MODE" = "dry-run" ]; then
        log_info "[DRY-RUN] Would copy: $source_dir -> $dest_dir"
    else
        cp -r "$source_dir"/* "$dest_dir"/
        log_success "Copied generated code to $name"
    fi
}

# Create commit and tag
create_commit_and_tag() {
    local repo_dir="$1"
    local version="$2"
    local message="$3"

    cd "$repo_dir"

    if [ "$MODE" = "manual" ]; then
        echo -n "Create commit and tag $version? (y/n): "
        read -r response
        if [[ ! "$response" =~ ^[Yy]$ ]]; then
            log_warning "Skipping commit and tag creation"
            return
        fi
    fi

    execute "git add -A" "Staging all changes"
    execute "git commit -m '$message'" "Creating commit"
    execute "git tag -a '$version' -m 'Release $version'" "Creating tag $version"

    log_success "Created commit and tag $version"
}

# Push to remote
push_to_remote() {
    local repo_dir="$1"
    local repo_name="$2"

    cd "$repo_dir"

    if [ "$MODE" = "auto" ]; then
        execute "git push origin HEAD" "Pushing commits to remote"
        execute "git push origin --tags" "Pushing tags to remote"
        log_success "Pushed $repo_name to remote"
    else
        log_info "Skipping push in $MODE mode (push manually when ready)"
    fi
}

# Release pango
release_pango() {
    log_info "=== Releasing pango ==="

    # Copy generated code
    copy_generated_code "$CODEGEN_DIR/target/pango" "$PANGO_DIR" "pango"

    # Determine next version
    local next_version=$(determine_next_version "$PANGO_DIR")

    if [ -z "$next_version" ] || [ "$next_version" = "$(cd $PANGO_DIR && git describe --tags --abbrev=0 2>/dev/null)" ]; then
        log_warning "No version bump needed for pango"
        return
    fi

    # Create commit and tag
    create_commit_and_tag "$PANGO_DIR" "$next_version" "chore(release): auto-generated $next_version"

    # Push to remote
    push_to_remote "$PANGO_DIR" "pango"

    log_success "Pango release completed: $next_version"
}

# Validate subcategories in docs
validate_subcategories() {
    local docs_dir="$TERRAFORM_DIR/docs"

    log_info "Validating subcategories in documentation..."

    if [ ! -d "$docs_dir" ]; then
        log_warning "Docs directory not found, skipping validation"
        return 0
    fi

    local missing=$(grep -r "subcategory: \$" "$docs_dir" 2>/dev/null || true)

    if [ -n "$missing" ]; then
        log_error "Found missing subcategories:"
        echo "$missing"

        if [ "$MODE" = "auto" ]; then
            log_error "Auto mode requires all subcategories to be present. Exiting."
            exit 1
        elif [ "$MODE" = "manual" ]; then
            echo -n "Continue anyway? (y/n): "
            read -r response
            if [[ ! "$response" =~ ^[Yy]$ ]]; then
                exit 1
            fi
        fi
        return 1
    else
        log_success "All subcategories present"
        return 0
    fi
}

# Release terraform provider
release_terraform_provider() {
    log_info "=== Releasing terraform-provider-panos ==="

    # Copy generated code
    copy_generated_code "$CODEGEN_DIR/target/terraform" "$TERRAFORM_DIR" "terraform-provider-panos"

    cd "$TERRAFORM_DIR"

    # Run gofix
    if [ -f "$GOFIX_SCRIPT" ]; then
        log_info "Running gofix to update pango dependency..."
        execute "bash '$GOFIX_SCRIPT'" "Running go-mod-replace.sh"
    else
        log_warning "gofix script not found at $GOFIX_SCRIPT, skipping"
    fi

    # Run go mod tidy
    execute "go mod tidy" "Running go mod tidy"

    # Generate docs
    log_info "Generating terraform docs..."
    execute "go generate ./..." "Running go generate"

    # Validate subcategories
    validate_subcategories

    # Determine next version
    local next_version=$(determine_next_version "$TERRAFORM_DIR")

    if [ -z "$next_version" ] || [ "$next_version" = "$(cd $TERRAFORM_DIR && git describe --tags --abbrev=0 2>/dev/null)" ]; then
        log_warning "No version bump needed for terraform-provider-panos"
        return
    fi

    # Create commit and tag
    create_commit_and_tag "$TERRAFORM_DIR" "$next_version" "chore(release): auto-generated $next_version"

    # Push to remote
    push_to_remote "$TERRAFORM_DIR" "terraform-provider-panos"

    log_success "Terraform provider release completed: $next_version"
}

# Display summary
display_summary() {
    log_info "=== Release Summary ==="
    echo ""
    echo "Mode: $MODE"
    echo ""
    echo "Pango:"
    echo "  Directory: $PANGO_DIR"
    echo "  Latest tag: $(cd $PANGO_DIR && git describe --tags --abbrev=0 2>/dev/null || echo 'none')"
    echo ""
    echo "Terraform Provider:"
    echo "  Directory: $TERRAFORM_DIR"
    echo "  Latest tag: $(cd $TERRAFORM_DIR && git describe --tags --abbrev=0 2>/dev/null || echo 'none')"
    echo ""

    if [ "$MODE" = "auto" ]; then
        log_success "Releases pushed to remote"
    else
        log_info "Changes are local only. Push manually when ready:"
        echo "  cd $PANGO_DIR && git push origin HEAD && git push origin --tags"
        echo "  cd $TERRAFORM_DIR && git push origin HEAD && git push origin --tags"
    fi
}

# Main function
main() {
    parse_args "$@"

    log_info "Running in $MODE mode..."
    echo ""

    # Check for required tools
    check_npx

    # Check all repositories are clean
    check_repo_clean "$CODEGEN_DIR"
    check_repo_clean "$PANGO_DIR"
    check_repo_clean "$TERRAFORM_DIR"

    # Run codegen
    run_codegen

    # Release pango
    release_pango

    # Release terraform provider
    release_terraform_provider

    # Display summary
    display_summary

    log_success "Release automation completed!"
}

# Run main
main "$@"
