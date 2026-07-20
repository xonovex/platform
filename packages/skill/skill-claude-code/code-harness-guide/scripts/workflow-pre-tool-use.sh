#!/usr/bin/env bash
set -euo pipefail

workflow_hook_executable="${XONOVEX_WORKFLOW_HOOK_EXECUTABLE:-}"

if [ -z "$workflow_hook_executable" ]; then
  printf '%s\n' 'XONOVEX_WORKFLOW_HOOK_EXECUTABLE is required.' >&2
  exit 2
fi
if [ ! -x "$workflow_hook_executable" ]; then
  printf 'Workflow hook executable is not executable: %s\n' "$workflow_hook_executable" >&2
  exit 2
fi

exec "$workflow_hook_executable"
