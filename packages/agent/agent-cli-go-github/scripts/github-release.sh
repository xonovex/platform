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

CHANGELOG=$(awk '/^## /{if(p)exit; p=1} p' "$CLI_DIR/CHANGELOG.md")

gh release create "$TAG" "$ASSETS_DIR"/* \
  --title "agent-cli-go v${VERSION}" \
  --notes "${CHANGELOG}"
