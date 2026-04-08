#!/usr/bin/env bash
#
# Generates markdown release notes from conventional commits in pan-os-codegen.
#
# Usage:
#   generate-release-notes.sh <version> [since-date]
#   generate-release-notes.sh <version> [--since-tag <tag>] [--since-date <date>] [--repo-url <url>]
#
# Options:
#   --since-tag   Use commits after this tag (preferred over date-based filtering)
#   --since-date  Only include commits after this date (fallback)
#   --repo-url    GitHub repository URL for links (auto-detected from git remote)
#
# Sections generated:
#   - Breaking Changes  (feat(MAJOR) or type! commits)
#   - New Resources     (feat(specs) commits, listed by resource name)
#   - Features          (feat commits, excluding specs)
#   - Bug Fixes         (fix commits)
#   - Performance       (perf commits)
#   - Reverts           (revert commits)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/conventional-commits.sh
source "$SCRIPT_DIR/lib/conventional-commits.sh"

# --- Argument parsing ---

VERSION=""
SINCE_TAG=""
SINCE_DATE=""
REPO_URL=""

# Support both old positional interface and new named flags
if [ $# -ge 1 ] && [[ "$1" != --* ]]; then
  VERSION="$1"
  shift
  # Check if second positional arg is a date (old interface) or a flag
  if [ $# -ge 1 ] && [[ "$1" != --* ]]; then
    SINCE_DATE="$1"
    shift
  fi
fi

while [ $# -gt 0 ]; do
  case "$1" in
    --since-tag)
      SINCE_TAG="$2"
      shift 2
      ;;
    --since-date)
      SINCE_DATE="$2"
      shift 2
      ;;
    --repo-url)
      REPO_URL="$2"
      shift 2
      ;;
    *)
      if [ -z "$VERSION" ]; then
        VERSION="$1"
        shift
      else
        echo "Unknown argument: $1" >&2
        exit 1
      fi
      ;;
  esac
done

if [ -z "$VERSION" ]; then
  echo "Usage: generate-release-notes.sh <version> [--since-tag <tag>] [--since-date <date>] [--repo-url <url>]" >&2
  exit 1
fi

if [ -z "$REPO_URL" ]; then
  REPO_URL=$(detect_repo_url)
fi

# --- Build git log range ---

LOG_RANGE=""
if [ -n "$SINCE_TAG" ]; then
  # Try codegen release tag format first
  CODEGEN_TAG="release/${SINCE_TAG}"
  if git rev-parse "$CODEGEN_TAG" >/dev/null 2>&1; then
    LOG_RANGE="${CODEGEN_TAG}..HEAD"
  elif git rev-parse "$SINCE_TAG" >/dev/null 2>&1; then
    LOG_RANGE="${SINCE_TAG}..HEAD"
  fi
fi

if [ -z "$LOG_RANGE" ] && [ -n "$SINCE_DATE" ]; then
  LOG_RANGE="--after=${SINCE_DATE} HEAD"
elif [ -z "$LOG_RANGE" ]; then
  LOG_RANGE="HEAD"
fi

# --- Collect and parse commits ---

# Use simple arrays for each section (bash 3.2 compatible)
BREAKING=()
NEW_RESOURCES=()
FEATS=()
FIXES=()
PERFS=()
REVERTS=()

while IFS= read -r line; do
  [ -z "$line" ] && continue

  # Split on ||| separator
  local_hash="${line%%|||*}"
  local_subject="${line#*|||}"

  parse_commit "$local_subject" "$local_hash"

  # Skip hidden types
  if ! is_visible "$CC_TYPE" && ! $CC_BREAKING; then
    continue
  fi

  # Format the entry
  entry=$(format_entry "$CC_MESSAGE" "$CC_PR" "$CC_HASH_FULL" "$REPO_URL")

  # Breaking changes (from feat(MAJOR) or ! suffix)
  if $CC_BREAKING; then
    BREAKING+=("- $entry")
    continue
  fi

  case "$CC_TYPE" in
    feat)
      if [[ "$CC_SCOPE" == "specs" || "$CC_SCOPE" == "spec" ]]; then
        # Extract resource name for New Resources section
        resource_name=$(extract_resource_name "$CC_MESSAGE")
        if [ -n "$resource_name" ]; then
          NEW_RESOURCES+=("- \`${resource_name}\` $(format_entry "" "$CC_PR" "$CC_HASH_FULL" "$REPO_URL")")
        else
          NEW_RESOURCES+=("- $entry")
        fi
      else
        if [ -n "$CC_SCOPE" ]; then
          FEATS+=("- **${CC_SCOPE}:** $entry")
        else
          FEATS+=("- $entry")
        fi
      fi
      ;;
    fix)
      if [ -n "$CC_SCOPE" ]; then
        FIXES+=("- **${CC_SCOPE}:** $entry")
      else
        FIXES+=("- $entry")
      fi
      ;;
    perf)
      if [ -n "$CC_SCOPE" ]; then
        PERFS+=("- **${CC_SCOPE}:** $entry")
      else
        PERFS+=("- $entry")
      fi
      ;;
    revert)
      REVERTS+=("- $entry")
      ;;
  esac
done < <(git log $LOG_RANGE --format="%H|||%s" --no-merges 2>/dev/null)

# Also scan commit bodies for BREAKING CHANGE footer
while IFS= read -r body_line; do
  [ -z "$body_line" ] && continue
  if [[ "$body_line" == "BREAKING CHANGE:"* ]]; then
    msg="${body_line#BREAKING CHANGE: }"
    BREAKING+=("- $msg")
  fi
done < <(git log $LOG_RANGE --format="%b" --no-merges 2>/dev/null)

# --- Output ---

echo "## What's Changed in ${VERSION}"
echo ""

if [ ${#BREAKING[@]} -gt 0 ]; then
  echo "### Breaking Changes"
  echo ""
  printf '%s\n' "${BREAKING[@]}"
  echo ""
fi

if [ ${#NEW_RESOURCES[@]} -gt 0 ]; then
  echo "### New Resources"
  echo ""
  printf '%s\n' "${NEW_RESOURCES[@]}"
  echo ""
fi

if [ ${#FEATS[@]} -gt 0 ]; then
  echo "### Features"
  echo ""
  printf '%s\n' "${FEATS[@]}"
  echo ""
fi

if [ ${#FIXES[@]} -gt 0 ]; then
  echo "### Bug Fixes"
  echo ""
  printf '%s\n' "${FIXES[@]}"
  echo ""
fi

if [ ${#PERFS[@]} -gt 0 ]; then
  echo "### Performance Improvements"
  echo ""
  printf '%s\n' "${PERFS[@]}"
  echo ""
fi

if [ ${#REVERTS[@]} -gt 0 ]; then
  echo "### Reverts"
  echo ""
  printf '%s\n' "${REVERTS[@]}"
  echo ""
fi

# Check if nothing was generated
if [ ${#BREAKING[@]} -eq 0 ] && [ ${#NEW_RESOURCES[@]} -eq 0 ] && \
   [ ${#FEATS[@]} -eq 0 ] && [ ${#FIXES[@]} -eq 0 ] && \
   [ ${#PERFS[@]} -eq 0 ] && [ ${#REVERTS[@]} -eq 0 ]; then
  echo "No notable changes."
  echo ""
fi
