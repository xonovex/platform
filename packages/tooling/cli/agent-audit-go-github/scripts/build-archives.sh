#!/bin/bash
set -euo pipefail

if [ "$#" -ne 2 ]; then
  echo "usage: $0 <cli-dir> <output-dir>" >&2
  exit 2
fi

CLI_DIR="$1"
OUTPUT_DIR="$2"

mkdir -p "$OUTPUT_DIR"

PLATFORMS=(darwin-arm64 darwin-x64 linux-arm64 linux-x64 win32-x64)

for platform in "${PLATFORMS[@]}"; do
  dir="$CLI_DIR/agent-audit-go-$platform/bin"
  if [ "$platform" = "win32-x64" ]; then
    (cd "$dir" && zip "$OUTPUT_DIR/agent-audit-go-$platform.zip" agent-audit-go.exe)
  else
    (cd "$dir" && tar -czvf "$OUTPUT_DIR/agent-audit-go-$platform.tar.gz" agent-audit-go)
  fi
done
