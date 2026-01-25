# idempotency: Idempotent Scripts

**Guideline:** Design scripts to be safely run multiple times without unintended side effects.

**Rationale:** Idempotent scripts are essential for automation, deployment, and configuration management. They allow safe retries after failures and prevent duplicate operations.

**Example:**

```sh
# ✅ Check before creating
create_directory() {
    local dir="$1"

    if [ -d "$dir" ]; then
        log "Directory exists: $dir"
        return 0
    fi

    mkdir -p "$dir"
    log "Created directory: $dir"
}

# ✅ Check before modifying
ensure_line_in_file() {
    local line="$1"
    local file="$2"

    if grep -qF "$line" "$file" 2>/dev/null; then
        log "Line already present: $file"
        return 0
    fi

    printf '%s\n' "$line" >> "$file"
    log "Added line to: $file"
}

# ✅ Atomic operations with temp files
update_config() {
    local config="$1"
    local temp
    temp="$(mktemp)"

    # Generate new config
    generate_config > "$temp"

    # Atomic replace
    mv "$temp" "$config"
}
```

**Techniques:**
- Check if resources exist before creating them
- Use conditional operations that skip when already done
- Use atomic operations with temporary files
- Log actions taken vs. skipped for observability
- Use `mkdir -p` and similar idempotent commands
- Verify state before and after modifications
