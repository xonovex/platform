# script-template: Basic Script Template

Starting point wiring together shebang, strict mode, defaults, logging/`die` helpers, dependency check, arg validation, and a `main`.

```sh
#!/usr/bin/env sh
# Usage: script.sh <arg1> [arg2]
set -eu

: "${VERBOSE:=0}"

log() { printf '%s\n' "$@"; }
error() { printf 'ERROR: %s\n' "$@" >&2; }
die() { error "$1"; exit "${2:-1}"; }

command -v required_tool >/dev/null 2>&1 || die "required_tool not found" 2
[ "$#" -ge 1 ] || die "Usage: $0 <name> [options]" 2

main() {
    name="$1"
    log "Processing: $name"
}

main "$@"
```
