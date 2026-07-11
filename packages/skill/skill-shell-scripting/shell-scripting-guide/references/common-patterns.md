# common-patterns: Common Patterns

Non-obvious idioms that avoid word splitting, dropped last lines, and leaks:

```sh
while IFS= read -r line; do process "$line"; done < "$input"   # IFS= and -r preserve whitespace/backslashes
temp_dir="$(mktemp -d)"; trap 'rm -rf "$temp_dir"' EXIT        # always trap temp cleanup
script_dir="$(cd "$(dirname "$0")" && pwd)"                    # absolute script dir
[ "$(id -u)" -eq 0 ] && log "running as root"
```
