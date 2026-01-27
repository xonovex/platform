#!/bin/bash
set -e

PLATFORMS_DIR="$1"
OUTPUT_DIR="$2"

mkdir -p "$OUTPUT_DIR"

for dir in "$PLATFORMS_DIR"/*/bin; do
  platform=$(basename "$(dirname "$dir")")
  if [ "$platform" = "win32-x64" ]; then
    (cd "$dir" && zip "$OUTPUT_DIR/agent-cli-go-$platform.zip" agent-cli-go.exe)
  else
    (cd "$dir" && tar -czvf "$OUTPUT_DIR/agent-cli-go-$platform.tar.gz" agent-cli-go)
  fi
done
