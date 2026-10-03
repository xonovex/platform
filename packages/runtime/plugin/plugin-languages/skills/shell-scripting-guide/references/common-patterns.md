# common-patterns: Common Patterns

Non-obvious idioms:

```sh
script_dir="$(cd "$(dirname "$0")" && pwd)"                    # absolute script dir
[ "$(id -u)" -eq 0 ] && log "running as root"                  # root check
```
