# functions: Function Best Practices

Declare every function-local variable `local` to avoid polluting global scope. `return` a status code (0 success, non-zero failure), never `exit` from a helper. A predicate function is just its test: `is_file() { [ -f "$1" ]; }`.

```sh
process_file() {
    local file="$1" output="$2"
    is_file "$file" || { error "not a file: $file"; return 1; }
    log "processed: $file -> $output"
}

if validate_input "$user_input"; then log "ok"; else die "invalid" "$?"; fi
```
