#!/bin/bash
set -euo pipefail

if [ "$#" -lt 2 ] || [ "$#" -gt 3 ]; then
  echo "usage: $0 <cli-dir> <assets-dir> [--dry-run]" >&2
  exit 2
fi

CLI_DIR="$1"
ASSETS_DIR="$2"
MODE="${3:-}"
[ -z "$MODE" ] || [ "$MODE" = "--dry-run" ] || {
  echo "unknown mode: $MODE" >&2
  exit 2
}

VERSION=$(node -p "require('$CLI_DIR/package.json').version")
TAG="agent-cli-go-v${VERSION}"

# Extract the exact release section and require it before publishing notes.
CHANGELOG=$(awk -v ver="$VERSION" '
  $1 == "##" && $2 == ver {p = 1}
  p && $1 == "##" && $2 != ver {exit}
  p
' "$CLI_DIR/CHANGELOG.md")
[ -n "$CHANGELOG" ] || {
  echo "CHANGELOG.md has no '## ${VERSION}' section" >&2
  exit 1
}

ASSETS=(
  "$ASSETS_DIR/agent-cli-go-darwin-arm64.tar.gz"
  "$ASSETS_DIR/agent-cli-go-darwin-x64.tar.gz"
  "$ASSETS_DIR/agent-cli-go-linux-arm64.tar.gz"
  "$ASSETS_DIR/agent-cli-go-linux-x64.tar.gz"
  "$ASSETS_DIR/agent-cli-go-win32-x64.zip"
)
for asset in "${ASSETS[@]}"; do
  [ -s "$asset" ] || {
    echo "release asset is missing or empty: $asset" >&2
    exit 1
  }
done

if [ "$MODE" = "--dry-run" ]; then
  echo "validated ${TAG}, changelog, and ${#ASSETS[@]} release assets"
  exit 0
fi

# Idempotent: a re-run (e.g. after a partial-publish failure where npm failed but the
# release was already cut) must not fail on an existing release. Skip if the tag is live.
if gh release view "$TAG" >/dev/null 2>&1; then
  echo "release ${TAG} already exists; skipping (bump the version to cut a new one)"
  exit 0
fi

gh release create "$TAG" "${ASSETS[@]}" \
  --title "agent-cli-go v${VERSION}" \
  --notes "${CHANGELOG}"
