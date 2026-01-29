# quoting: Quoting Variables and Expansions

**Guideline:** Always quote variable expansions with double quotes to prevent word splitting and glob expansion.

**Rationale:** Unquoted variables are subject to word splitting on whitespace and pathname expansion, leading to subtle bugs when filenames contain spaces or special characters.

**Example:**

```sh
# ✅ Always quote variable expansions
name="John Doe"
log "Hello, $name"
file="/path/to/my file.txt"
cat "$file"

# ❌ Unquoted variables cause word splitting
cat $file  # Tries to cat "/path/to/my" and "file.txt" separately!

# ✅ Quote arrays in loops
for file in "$@"; do
    process "$file"
done

# ❌ Unquoted $@ splits on spaces
for file in $@; do  # Wrong!
    process "$file"
done

# ✅ Use "${var}" for clarity
config_file="${CONFIG_DIR}/app.conf"

# ✅ Quote command substitution
current_date="$(date +%Y-%m-%d)"
files="$(find . -name '*.txt')"
```

**Techniques:**

- Quote all variable expansions: `"$var"`
- Quote command substitutions: `"$(command)"`
- Quote `"$@"` in loops to preserve arguments with spaces
- Use `"${var}"` for clarity in complex expansions
- Never use unquoted variables in file operations
