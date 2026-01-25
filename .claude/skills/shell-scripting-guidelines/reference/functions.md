# functions: Function Best Practices

**Guideline:** Write small, focused functions with local variables and meaningful return codes.

**Rationale:** Small functions improve readability, testability, and reusability. Local variables prevent variable pollution, and return codes enable proper error handling.

**Example:**

```sh
# ✅ Small focused functions
is_file() {
    [ -f "$1" ]
}

is_dir() {
    [ -d "$1" ]
}

file_exists() {
    [ -e "$1" ]
}

# ✅ Use local variables
process_file() {
    local file="$1"
    local output="$2"

    if ! is_file "$file"; then
        error "Not a file: $file"
        return 1
    fi

    # Process file...
    log "Processed: $file -> $output"
}

# ✅ Return status codes
validate_input() {
    local input="$1"

    if [ -z "$input" ]; then
        return 1
    fi

    if [ "${#input}" -lt 3 ]; then
        return 2
    fi

    return 0
}

# Usage
if validate_input "$user_input"; then
    log "Valid input"
else
    die "Invalid input" "$?"
fi
```

**Techniques:**
- Keep functions focused on a single task
- Declare all function variables as `local`
- Return 0 for success, non-zero for failure
- Use descriptive function names (verb_noun format)
- Validate function arguments before use
- Use return codes, not exit, within functions
