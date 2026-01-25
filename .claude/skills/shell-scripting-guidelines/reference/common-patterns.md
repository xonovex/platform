# common-patterns: Common Patterns

**Guideline:** Use well-tested patterns for common operations like reading files, processing output, and managing temporary resources.

**Rationale:** Standard patterns are proven to handle edge cases correctly and avoid common pitfalls like word splitting, missing the last line, or resource leaks.

**Example:**

```sh
# ✅ Read file line by line
while IFS= read -r line; do
    process_line "$line"
done < "$input_file"

# ✅ Process command output
find . -name '*.txt' | while IFS= read -r file; do
    process_file "$file"
done

# ✅ Temporary directory
temp_dir="$(mktemp -d)"
trap 'rm -rf "$temp_dir"' EXIT

# ✅ Check if running as root
if [ "$(id -u)" -eq 0 ]; then
    log "Running as root"
fi

# ✅ Get script directory
script_dir="$(cd "$(dirname "$0")" && pwd)"

# ✅ Check exit status
if command arg1 arg2; then
    log "Command succeeded"
else
    error "Command failed with exit code: $?"
fi
```

**Techniques:**
- Use `while IFS= read -r line` to read files line by line
- Create temp files/dirs with `mktemp` and always trap cleanup
- Use command substitution in conditionals for proper error handling
- Get script directory with `cd "$(dirname "$0")" && pwd`
- Check root with `[ "$(id -u)" -eq 0 ]`
- Preserve IFS in read loops to handle whitespace correctly
