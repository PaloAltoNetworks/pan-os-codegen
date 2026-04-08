#!/usr/bin/env bash
#
# Shared library for parsing conventional commits and formatting release notes.
#
# Usage:
#   source "$(dirname "$0")/lib/conventional-commits.sh"
#
# Provides:
#   parse_commit <subject> <hash>    — sets CC_TYPE, CC_SCOPE, CC_BREAKING, CC_MESSAGE, CC_PR, CC_HASH
#   format_entry <message> <pr> <hash_full> <repo_url> — returns formatted markdown line with links
#   extract_resource_name <message>  — extracts panos_* resource name from specs commit messages
#   is_visible <type>                — returns 0 if type should appear in release notes

# Output variables set by parse_commit
CC_TYPE=""
CC_SCOPE=""
CC_BREAKING=false
CC_MESSAGE=""
CC_PR=""
CC_HASH=""
CC_HASH_FULL=""

# parse_commit <subject> <hash>
#
# Parses a conventional commit subject line into its components.
# Sets global variables: CC_TYPE, CC_SCOPE, CC_BREAKING, CC_MESSAGE, CC_PR, CC_HASH, CC_HASH_FULL
parse_commit() {
  local subject="$1"
  local hash="$2"

  CC_TYPE=""
  CC_SCOPE=""
  CC_BREAKING=false
  CC_MESSAGE=""
  CC_PR=""
  CC_HASH="${hash:0:7}"
  CC_HASH_FULL="$hash"

  # Extract PR number from subject: (#123)
  # Store regex in variable for bash 3.2 compatibility
  local pr_re='\(#([0-9]+)\)'
  if [[ "$subject" =~ $pr_re ]]; then
    CC_PR="${BASH_REMATCH[1]}"
  fi

  # Parse conventional commit format: type(scope)!: message
  # Store regex in variable for bash 3.2 compatibility
  local cc_re='^([a-zA-Z]+)(\(([^)]+)\))?(!)?: (.+)$'
  if [[ "$subject" =~ $cc_re ]]; then
    CC_TYPE="${BASH_REMATCH[1]}"
    # Lowercase the type (compatible with bash 3.2)
    CC_TYPE=$(echo "$CC_TYPE" | tr '[:upper:]' '[:lower:]')
    CC_SCOPE="${BASH_REMATCH[3]}"
    [[ "${BASH_REMATCH[4]}" == "!" ]] && CC_BREAKING=true
    CC_MESSAGE="${BASH_REMATCH[5]}"
  else
    CC_TYPE="other"
    CC_SCOPE=""
    CC_MESSAGE="$subject"
  fi

  # feat(MAJOR) is treated as breaking
  if [[ "$CC_TYPE" == "feat" && "$CC_SCOPE" == "MAJOR" ]]; then
    CC_BREAKING=true
  fi
}

# format_entry <message> <pr> <hash_full> <repo_url>
#
# Returns a formatted markdown entry with PR and commit links.
# The message should NOT include the PR reference — it will be appended as a link.
format_entry() {
  local message="$1"
  local pr="$2"
  local hash_full="$3"
  local repo_url="$4"

  local hash="${hash_full:0:7}"
  local line="$message"

  # Strip trailing PR reference from message if present, since we add it as a link
  if [ -n "$pr" ]; then
    line="${line% (#$pr)}"
    line="${line% (#${pr})}"
  fi

  # Use space separator only when line is non-empty
  local sep=""
  [ -n "$line" ] && sep=" "

  # Append PR link
  if [ -n "$pr" ] && [ -n "$repo_url" ]; then
    line="${line}${sep}([#${pr}](${repo_url}/issues/${pr}))"
    sep=" "
  elif [ -n "$pr" ]; then
    line="${line}${sep}(#${pr})"
    sep=" "
  fi

  # Append commit hash link
  if [ -n "$hash_full" ] && [ -n "$repo_url" ]; then
    line="${line}${sep}([${hash}](${repo_url}/commit/${hash_full}))"
  elif [ -n "$hash_full" ]; then
    line="${line}${sep}(${hash})"
  fi

  echo "$line"
}

# extract_resource_name <message>
#
# Extracts a panos_* resource name from a specs commit message.
# Returns empty string if no resource name found.
extract_resource_name() {
  local message="$1"
  local resource=""

  # Try to match panos_* pattern anywhere in the message
  local res_re='(panos_[a-z_]+)'
  if [[ "$message" =~ $res_re ]]; then
    resource="${BASH_REMATCH[1]}"
  fi

  echo "$resource"
}

# is_visible <type>
#
# Returns 0 if the type should appear in release notes, 1 if hidden.
# Visible types: feat, fix, perf, revert
# Hidden types: docs, style, chore, refactor, build, ci, test, other
is_visible() {
  local type="$1"
  case "$type" in
    feat|fix|perf|revert) return 0 ;;
    *) return 1 ;;
  esac
}

# detect_repo_url
#
# Detects the GitHub repository URL from git remote origin.
detect_repo_url() {
  local url
  url=$(git remote get-url origin 2>/dev/null || echo "")
  # Convert SSH URL to HTTPS
  url="${url/git@github.com:/https://github.com/}"
  # Strip .git suffix
  url="${url%.git}"
  echo "$url"
}
