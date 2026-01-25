# error-handling: Error Handling

**Guideline:** Provide clear error messages with context, validate operations, and implement cleanup handlers.

**Rationale:** Good error handling makes scripts easier to debug and prevents resource leaks. Clear error messages help users understand and fix problems quickly.

**Example:**

```sh
# ✅ Clear error messages with context
die() {
    printf 'ERROR: %s\n' "$1" >&2
    exit "${2:-1}"
}

# ✅ Check command existence
require_command() {
    local cmd="$1"
    if ! command -v "$cmd" >/dev/null 2>&1; then
        die "Required command not found: $cmd" 127
    fi
}

# ✅ Validate file operations
read_config() {
    local config="$1"

    if [ ! -f "$config" ]; then
        die "Config file not found: $config" 2
    fi

    if [ ! -r "$config" ]; then
        die "Config file not readable: $config" 2
    fi

    cat "$config"
}

# ✅ Cleanup on exit
cleanup() {
    local exit_code=$?
    if [ -n "${temp_file:-}" ]; then
        rm -f "$temp_file"
    fi
    exit "$exit_code"
}

trap cleanup EXIT INT TERM

# Create temp file
temp_file="$(mktemp)"
# Use temp_file...
# Cleanup happens automatically
```

**Techniques:**
- Create a `die()` function for fatal errors with exit codes
- Check command existence with `command -v`
- Validate file operations before performing them
- Use trap to ensure cleanup on exit, interrupt, or termination
- Include context in error messages (filenames, expected values)
- Use appropriate exit codes (0=success, 1=general error, 2=usage error, 127=command not found)
