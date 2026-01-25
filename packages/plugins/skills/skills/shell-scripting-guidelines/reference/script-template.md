# script-template: Basic Script Template

**Guideline:** Start all shell scripts with a standard template including shebang, strict mode, logging functions, and argument validation.

**Rationale:** A consistent template ensures scripts follow safety best practices from the start, reducing common errors and improving maintainability.

**Example:**

```sh
#!/usr/bin/env sh
# Description: Brief description of what this script does
# Usage: script.sh <arg1> [arg2]

set -eu  # Exit on error, error on undefined variable
# Note: pipefail not in POSIX, use only if targeting bash

# Default values
: "${TMPDIR:=/tmp}"
: "${VERBOSE:=0}"

# Logging functions
log() {
    printf '%s\n' "$@"
}

log_verbose() {
    if [ "$VERBOSE" = "1" ]; then
        printf '[VERBOSE] %s\n' "$@" >&2
    fi
}

error() {
    printf 'ERROR: %s\n' "$@" >&2
}

die() {
    error "$1"
    exit "${2:-1}"
}

# Check requirements
command -v required_tool >/dev/null 2>&1 || die "required_tool not found" 2

# Validate arguments
if [ "$#" -lt 1 ]; then
    die "Usage: $0 <name> [options]" 2
fi

# Main logic
main() {
    local name="$1"
    log "Processing: $name"

    # Do work...

    log "Done"
}

# Run main
main "$@"
```

**Techniques:**
- Begin with `#!/usr/bin/env sh` for portability
- Add script description and usage in comments
- Set strict mode with `set -eu`
- Define logging and error functions
- Check for required dependencies
- Validate arguments before processing
- Implement main logic in a main() function
