#!/bin/bash
#
# Generates markdown release notes from conventional commits in pan-os-codegen.
#
# Usage:
#   generate-release-notes.sh <version> [since-date]
#
# If since-date is provided, only includes commits after that date.
# Otherwise includes all commits.

set -euo pipefail

VERSION="${1:?Usage: generate-release-notes.sh <version> [since-date]}"
SINCE_DATE="${2:-}"

LOG_ARGS="--format=%s"
if [ -n "$SINCE_DATE" ]; then
  LOG_ARGS="$LOG_ARGS --after=$SINCE_DATE"
fi

# Collect commits into arrays by type
FEATS=()
FIXES=()
BREAKING=()

while IFS= read -r subject; do
  [ -z "$subject" ] && continue

  # feat(MAJOR) -> breaking
  if echo "$subject" | grep -qiE '^feat\(MAJOR\)'; then
    msg=$(echo "$subject" | sed 's/^feat(MAJOR)[!]*: //')
    BREAKING+=("- $msg")
    continue
  fi

  # feat -> feature
  if echo "$subject" | grep -qE '^feat(\(|:)'; then
    scope=$(echo "$subject" | sed -n 's/^feat(\([^)]*\))[!]*:.*/\1/p')
    msg=$(echo "$subject" | sed 's/^feat([^)]*)[!]*: //' | sed 's/^feat[!]*: //')
    if [ -n "$scope" ]; then
      FEATS+=("- **$scope**: $msg")
    else
      FEATS+=("- $msg")
    fi
    continue
  fi

  # fix -> bug fix
  if echo "$subject" | grep -qE '^fix(\(|:)'; then
    scope=$(echo "$subject" | sed -n 's/^fix(\([^)]*\))[!]*:.*/\1/p')
    msg=$(echo "$subject" | sed 's/^fix([^)]*)[!]*: //' | sed 's/^fix[!]*: //')
    if [ -n "$scope" ]; then
      FIXES+=("- **$scope**: $msg")
    else
      FIXES+=("- $msg")
    fi
    continue
  fi

  # Skip chore, docs, ci, test, refactor, style, build — internal commits
done < <(git log $LOG_ARGS HEAD 2>/dev/null)

# Also scan commit bodies for BREAKING CHANGE
while IFS= read -r body_line; do
  if echo "$body_line" | grep -q 'BREAKING CHANGE:'; then
    msg=$(echo "$body_line" | sed 's/BREAKING CHANGE: //')
    BREAKING+=("- $msg")
  fi
done < <(git log ${SINCE_DATE:+--after=$SINCE_DATE} --format="%b" HEAD 2>/dev/null)

# Output
echo "## What's Changed in ${VERSION}"
echo ""

if [ ${#BREAKING[@]} -gt 0 ]; then
  echo "### Breaking Changes"
  echo ""
  printf '%s\n' "${BREAKING[@]}"
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

if [ ${#BREAKING[@]} -eq 0 ] && [ ${#FEATS[@]} -eq 0 ] && [ ${#FIXES[@]} -eq 0 ]; then
  echo "No notable changes."
  echo ""
fi
