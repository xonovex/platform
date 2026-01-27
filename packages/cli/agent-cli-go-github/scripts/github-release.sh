#!/bin/bash
set -e

CLI_DIR="$1"
ASSETS_DIR="$2"

VERSION=$(node -p "require('$CLI_DIR/package.json').version")
CHANGELOG=$(awk '/^## /{if(p)exit; p=1} p' "$CLI_DIR/CHANGELOG.md")

gh release create "agent-cli-go-v${VERSION}" "$ASSETS_DIR"/* \
  --title "agent-cli-go v${VERSION}" \
  --notes "${CHANGELOG}"
