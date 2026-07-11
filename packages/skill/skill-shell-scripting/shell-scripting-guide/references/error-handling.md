# error-handling: Error Handling

Fatal errors go through a `die()` helper that prints to stderr and exits with a meaningful code: 0 success, 1 general, 2 usage/validation, 127 command not found. Validate file readability/existence before use. Register a `trap` so cleanup runs on EXIT/INT/TERM; capture `$?` first so the original exit code survives.

```sh
die() { printf 'ERROR: %s\n' "$1" >&2; exit "${2:-1}"; }

require_command() { command -v "$1" >/dev/null 2>&1 || die "not found: $1" 127; }

cleanup() { code=$?; [ -n "${temp_file:-}" ] && rm -f "$temp_file"; exit "$code"; }
trap cleanup EXIT INT TERM
temp_file="$(mktemp)"
```
