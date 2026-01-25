# posix-compatibility: POSIX Compatibility

**Guideline:** Use POSIX-compliant syntax for maximum portability across different shells and systems.

**Rationale:** POSIX-compliant scripts run on any POSIX shell (sh, dash, bash, zsh, etc.), ensuring scripts work reliably across different systems without modification.

**Example:**

```sh
# ✅ Use POSIX test constructs
if [ -f "$file" ]; then
    log "File exists"
fi

# ❌ Avoid bash-specific [[
if [[ -f $file ]]; then  # Not POSIX
    log "File exists"
fi

# ✅ POSIX string comparison
if [ "$status" = "success" ]; then
    log "Success"
fi

# ✅ POSIX numeric comparison
if [ "$count" -gt 10 ]; then
    log "Count exceeds 10"
fi

# ✅ POSIX command existence check
if command -v git >/dev/null 2>&1; then
    log "git is available"
fi

# ❌ Avoid type/which
if type git >/dev/null 2>&1; then  # Less portable
    log "git is available"
fi

# ✅ POSIX-compatible string operations
string="hello world"
# Use parameter expansion instead of bash ${var^^}
upper="$(printf '%s' "$string" | tr '[:lower:]' '[:upper:]')"
```

**Techniques:**
- Use `[ ]` instead of bash-specific `[[ ]]`
- Use `=` for string comparison, not `==`
- Use `-gt`, `-lt`, etc. for numeric comparisons
- Use `command -v` instead of `which` or `type`
- Use `tr` and `cut` instead of bash string operations
- Avoid bash arrays, use positional parameters or temp files
- Test scripts with `sh` (dash) not just bash
