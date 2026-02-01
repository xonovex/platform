#!/bin/bash
set -e

CLI_DIR="$1"
OUTPUT_DIR="$2"

mkdir -p "$OUTPUT_DIR"

PLATFORMS="darwin-arm64 darwin-x64 linux-arm64 linux-x64 win32-x64"

for platform in $PLATFORMS; do
  dir="$CLI_DIR/agent-cli-go/platforms/agent-cli-go-$platform/bin"
  if [ "$platform" = "win32-x64" ]; then
    (cd "$dir" && zip "$OUTPUT_DIR/agent-cli-go-$platform.zip" agent-cli-go.exe)
  else
    (cd "$dir" && tar -czvf "$OUTPUT_DIR/agent-cli-go-$platform.tar.gz" agent-cli-go)
  fi
done
