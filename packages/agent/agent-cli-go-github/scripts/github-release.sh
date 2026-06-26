#!/bin/bash
set -e

CLI_DIR="$1"
ASSETS_DIR="$2"

VERSION=$(node -p "require('$CLI_DIR/package.json').version")
TAG="agent-cli-go-v${VERSION}"

# Idempotent: a re-run (e.g. after a partial-publish failure where npm failed but the
# release was already cut) must not fail on an existing release. Skip if the tag is live.
if gh release view "$TAG" >/dev/null 2>&1; then
  echo "release ${TAG} already exists; skipping (bump the version to cut a new one)"
  exit 0
fi

# Extract the section for THIS version (not just the top entry), so a missing changelog
# entry publishes no notes rather than the previous version's. Fail loudly if absent.
CHANGELOG=$(awk -v ver="$VERSION" '
  $1 == "##" && $2 == ver {p = 1}
  p && $1 == "##" && $2 != ver {exit}
  p
' "$CLI_DIR/CHANGELOG.md")
[ -n "$CHANGELOG" ] || {
  echo "CHANGELOG.md has no '## ${VERSION}' section" >&2
  exit 1
}

gh release create "$TAG" "$ASSETS_DIR"/* \
  --title "agent-cli-go v${VERSION}" \
  --notes "${CHANGELOG}"
