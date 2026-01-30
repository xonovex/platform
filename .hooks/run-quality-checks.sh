#!/usr/bin/env bash
set -Eeuo pipefail

# Generic quality checks runner using Moon
# Runs lint-fix, typecheck, test, build on affected files
# Returns structured output for callers to parse

usage() {
  cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Options:
  --project-dir DIR    Project directory (default: current dir)
  --log-file FILE      Log file path (default: temp file, deleted on success)
  --targets TARGETS    Comma-separated moon targets (default: :lint-fix,:typecheck,:test,:build)
  --affected           Run on affected files (default)
  --staged             Run on staged files only (for pre-commit)
  --no-bail            Continue running after failures
  --concurrency N      Max concurrent tasks
  --timeout SECS       Timeout in seconds (default: 600)
  -h, --help           Show this help

Exit codes:
  0 = All checks passed
  1 = Checks failed
  2 = Timeout
  3 = Checks passed but files were modified (lint-fix)
EOF
}

# Defaults
PROJECT_DIR="$(pwd)"
LOG_FILE=""
TARGETS=":lint-fix,:typecheck,:test,:build"
NO_BAIL=0
CONCURRENCY=""
TIMEOUT=600
STATUS_MODE="affected"  # "affected" or "staged"

# Parse arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    --project-dir)   PROJECT_DIR="$2"; shift 2 ;;
    --log-file)      LOG_FILE="$2"; shift 2 ;;
    --targets)       TARGETS="$2"; shift 2 ;;
    --affected)      STATUS_MODE="affected"; shift ;;
    --staged)        STATUS_MODE="staged"; shift ;;
    --no-bail)       NO_BAIL=1; shift ;;
    --concurrency)   CONCURRENCY="$2"; shift 2 ;;
    --timeout)       TIMEOUT="$2"; shift 2 ;;
    -h|--help)       usage; exit 0 ;;
    *)               echo "Unknown option: $1" >&2; usage >&2; exit 1 ;;
  esac
done

# Create temp log file if not specified
temp_log=0
if [[ -z "$LOG_FILE" ]]; then
  LOG_FILE="$(mktemp)"
  temp_log=1
fi

cleanup_temp_log() {
  if (( temp_log )) && [[ -f "$LOG_FILE" ]]; then
    rm -f "$LOG_FILE"
  fi
}

# Pick moon executable (prefer local, then global, then npx)
if [[ -x "$PROJECT_DIR/node_modules/.bin/moon" ]]; then
  MOON=( "$PROJECT_DIR/node_modules/.bin/moon" )
elif command -v moon >/dev/null 2>&1; then
  MOON=( moon )
else
  MOON=( npx --yes moon )
fi

# Build target array from comma-separated string
IFS=',' read -ra TARGET_ARRAY <<< "$TARGETS"

# Build moon command
MOON_CMD=( "${MOON[@]}" run "${TARGET_ARRAY[@]}" )

# Add status mode (affected vs staged)
if [[ "$STATUS_MODE" == "staged" ]]; then
  MOON_CMD+=( --affected --status=staged )
else
  MOON_CMD+=( --affected )
fi

# Suppress progress output - only show on error
MOON_CMD+=( --log error )

if [[ -n "$CONCURRENCY" ]]; then
  MOON_CMD+=( --concurrency "$CONCURRENCY" )
fi

if (( NO_BAIL )); then
  MOON_CMD+=( --no-bail )
fi

# Detect file modifications (lint:fix might modify files)
git_ok=0
before_tmp=""
after_tmp=""

cleanup() {
  [[ -n "${before_tmp:-}" ]] && rm -f -- "$before_tmp" || true
  [[ -n "${after_tmp:-}" ]] && rm -f -- "$after_tmp" || true
}
trap cleanup EXIT

if command -v git >/dev/null 2>&1 && git -C "$PROJECT_DIR" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  git_ok=1
  before_tmp="$(mktemp)"
  after_tmp="$(mktemp)"
  git -C "$PROJECT_DIR" diff --name-only >"$before_tmp" 2>/dev/null || true
fi

# Run quality checks
set +e
FULL_COMMAND="cd \"$PROJECT_DIR\" && ${MOON_CMD[*]@Q}"
timeout "$TIMEOUT" bash -c "$FULL_COMMAND" >"$LOG_FILE" 2>&1
cmd_status=$?
set -e

# Check for timeout
if [[ $cmd_status -eq 124 ]]; then
  echo "TIMEOUT:$TIMEOUT" >&2
  echo "LOG:$LOG_FILE" >&2
  exit 2
fi

# Check for modified files
changed_files=""
if (( git_ok )); then
  git -C "$PROJECT_DIR" diff --name-only >"$after_tmp" 2>/dev/null || true
  changed_files="$(comm -13 <(sort -u "$before_tmp") <(sort -u "$after_tmp") | head -n 50 || true)"
fi

# Success case
if [[ "$cmd_status" -eq 0 ]]; then
  if [[ -n "$changed_files" ]]; then
    echo "MODIFIED_FILES:$changed_files" >&2
    echo "LOG:$LOG_FILE" >&2
    exit 3
  fi
  cleanup_temp_log
  echo "OK"
  exit 0
fi

# Failure case - extract useful context from log

# Extract moon's summary of which task failed (appears after run_failed error)
failed_task_summary="$(
  tail -100 "$LOG_FILE" \
    | sed -n '/Error: task_runner::run_failed/,$ p' \
    | head -n 10 \
    || true
)"

# Extract actual error lines (TypeScript errors, npm errors, build failures)
# These are the most useful for debugging
actual_errors="$(
  grep -E '(error TS[0-9]+:|: error:|npm error command failed|Process .* failed:|Could not resolve|Build failed|FAIL [^r]|AssertionError|Expected .* but|Cannot find module)' "$LOG_FILE" \
    | grep -v -E '(\[ERROR\].*test \||should .* FAIL|record FAIL)' \
    | head -n 15 \
    || true
)"

# If no specific errors found, get context around the failure
if [[ -z "$actual_errors" ]]; then
  actual_errors="$(
    tail -200 "$LOG_FILE" \
      | grep -B 5 'Error: task_runner::run_failed' \
      | head -20 \
      || true
  )"
fi

moon_failure_context="$actual_errors"

# Output structured failure info
echo "FAILED:$cmd_status" >&2
[[ -n "$failed_task_summary" ]] && echo "TASK_SUMMARY:$failed_task_summary" >&2
[[ -n "$moon_failure_context" ]] && echo "ERROR_CONTEXT:$moon_failure_context" >&2
[[ -n "$changed_files" ]] && echo "MODIFIED_FILES:$changed_files" >&2
echo "LOG:$LOG_FILE" >&2
echo "COMMAND:${MOON_CMD[*]}" >&2

exit 1
