#!/bin/bash
#
# Determines the next version for terraform-provider-panos based on
# conventional commits in pan-os-codegen since the last provider release.
#
# Custom release rules (non-standard semver):
#   feat(MAJOR): ...           -> major bump
#   BREAKING CHANGE in footer  -> minor bump (not major!)
#   feat: ...                  -> patch bump (not minor!)
#   fix: ...                   -> patch bump
#
# Usage:
#   determine-version.sh                              # auto-detect from local repos
#   determine-version.sh --provider-dir <path>        # specify provider repo path
#   determine-version.sh --last-tag <tag>             # specify last tag directly (for CI)

set -euo pipefail

PROVIDER_DIR=""
LAST_TAG=""

while [[ $# -gt 0 ]]; do
  case $1 in
    --provider-dir) PROVIDER_DIR="$2"; shift 2 ;;
    --last-tag) LAST_TAG="$2"; shift 2 ;;
    *) echo "Unknown argument: $1" >&2; exit 1 ;;
  esac
done

# Resolve the last tag
if [ -z "$LAST_TAG" ]; then
  if [ -n "$PROVIDER_DIR" ] && [ -d "$PROVIDER_DIR/.git" ]; then
    LAST_TAG=$(cd "$PROVIDER_DIR" && git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
  elif command -v gh &>/dev/null; then
    LAST_TAG=$(gh release view --repo PaloAltoNetworks/terraform-provider-panos --json tagName -q '.tagName' 2>/dev/null || echo "v0.0.0")
  else
    echo "Error: cannot determine last tag. Provide --last-tag or --provider-dir." >&2
    exit 1
  fi
fi

CURRENT_VERSION="${LAST_TAG#v}"

# Determine the anchor date for codegen commits
if [ "$LAST_TAG" = "v0.0.0" ]; then
  SINCE_FLAG=""
else
  TAG_DATE=""
  if [ -n "$PROVIDER_DIR" ] && [ -d "$PROVIDER_DIR/.git" ]; then
    TAG_DATE=$(cd "$PROVIDER_DIR" && git log -1 --format="%aI" "$LAST_TAG" 2>/dev/null || echo "")
  fi
  if [ -z "$TAG_DATE" ] && command -v gh &>/dev/null; then
    TAG_DATE=$(gh release view "$LAST_TAG" --repo PaloAltoNetworks/terraform-provider-panos --json publishedAt -q '.publishedAt' 2>/dev/null || echo "")
  fi
  if [ -n "$TAG_DATE" ]; then
    SINCE_FLAG="--after=$TAG_DATE"
  else
    echo "Warning: could not determine tag date, scanning all commits" >&2
    SINCE_FLAG=""
  fi
fi

IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT_VERSION"

BUMP="none"

while IFS= read -r line; do
  [ -z "$line" ] && continue

  # Highest priority: feat(MAJOR) -> major bump
  if echo "$line" | grep -qiE '^feat\(MAJOR\)'; then
    BUMP="major"
    break
  fi

  # BREAKING CHANGE in commit body -> minor bump
  if echo "$line" | grep -q 'BREAKING CHANGE'; then
    if [ "$BUMP" != "major" ]; then
      BUMP="minor"
    fi
    continue
  fi

  # feat (but not feat(MAJOR)) -> patch bump
  if echo "$line" | grep -qE '^feat(\(|:)' && ! echo "$line" | grep -qiE '^feat\(MAJOR\)'; then
    if [ "$BUMP" = "none" ]; then
      BUMP="patch"
    fi
    continue
  fi

  # fix -> patch bump
  if echo "$line" | grep -qE '^fix(\(|:)'; then
    if [ "$BUMP" = "none" ]; then
      BUMP="patch"
    fi
    continue
  fi
done <<< "$(git log $SINCE_FLAG --format="%s%n%b" HEAD 2>/dev/null)"

case $BUMP in
  major)
    MAJOR=$((MAJOR + 1))
    MINOR=0
    PATCH=0
    ;;
  minor)
    MINOR=$((MINOR + 1))
    PATCH=0
    ;;
  patch)
    PATCH=$((PATCH + 1))
    ;;
  none)
    echo "NO_BUMP"
    exit 0
    ;;
esac

echo "v${MAJOR}.${MINOR}.${PATCH}"
