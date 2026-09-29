#!/bin/bash
#
# Runs govulncheck against a Go module and fails if any vulnerability is
# reachable from the module's code, unless it is listed in the allowlist.
#
# Only reachable findings (where govulncheck found a call path to the
# vulnerable symbol) fail the check. Findings that only match at the module
# or package level are reported but do not fail.
#
# Allowlist format (default: .govulncheck-ignore in the repo root), one
# OSV ID per line, "#" starts a comment:
#   GO-2026-6443  # no upstream fix released yet, revisit on grpc v1.85.0
#
# Usage:
#   govulncheck.sh <module-dir> [--ignore-file <path>]

set -euo pipefail

GOVULNCHECK_VERSION="v1.8.0"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
IGNORE_FILE="${SCRIPT_DIR}/../.govulncheck-ignore"
MODULE_DIR=""

while [[ $# -gt 0 ]]; do
  case $1 in
    --ignore-file) IGNORE_FILE="$2"; shift 2 ;;
    *) MODULE_DIR="$1"; shift ;;
  esac
done

if [ -z "$MODULE_DIR" ]; then
  echo "Usage: $0 <module-dir> [--ignore-file <path>]" >&2
  exit 2
fi

IGNORED=""
if [ -f "$IGNORE_FILE" ]; then
  IGNORED=$(sed 's/#.*//' "$IGNORE_FILE" | tr -s '[:space:]' '\n' | grep -v '^$' || true)
fi

REPORT=$(cd "$MODULE_DIR" && go run "golang.org/x/vuln/cmd/govulncheck@${GOVULNCHECK_VERSION}" -format json ./...)

# A finding is reachable when the innermost trace frame names a function.
REACHABLE=$(echo "$REPORT" | jq -r 'select(.finding) | .finding | select(.trace[0].function) | .osv' | sort -u)
UNREACHABLE=$(echo "$REPORT" | jq -r 'select(.finding) | .finding | select(.trace[0].function | not) | .osv' | sort -u)
UNREACHABLE=$(comm -23 <(echo "$UNREACHABLE") <(echo "$REACHABLE") | grep -v '^$' || true)

FAILED=""
for id in $REACHABLE; do
  if echo "$IGNORED" | grep -qx "$id"; then
    echo "::warning::${MODULE_DIR}: reachable vulnerability ${id} is allowlisted in $(basename "$IGNORE_FILE")"
  else
    echo "::error::${MODULE_DIR}: reachable vulnerability ${id} (https://pkg.go.dev/vuln/${id})"
    FAILED="yes"
  fi
done

for id in $UNREACHABLE; do
  echo "::notice::${MODULE_DIR}: ${id} is present in a dependency but not reachable (https://pkg.go.dev/vuln/${id})"
done

if [ -n "$FAILED" ]; then
  echo "Run 'go run golang.org/x/vuln/cmd/govulncheck@${GOVULNCHECK_VERSION} ./...' in ${MODULE_DIR} for call stacks." >&2
  exit 1
fi

echo "${MODULE_DIR}: no reachable vulnerabilities outside the allowlist"
