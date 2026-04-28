# parameter-expansion: Default Values and Parameter Expansion

**Guideline:** Use parameter expansion to set default values, require variables, and perform string operations.

**Rationale:** Parameter expansion provides safe, portable ways to handle optional and required variables without external commands, improving script robustness and performance.

**Example:**

```sh
# ✅ Default value if unset or empty
: "${CONFIG_FILE:=/etc/app/config.conf}"
: "${PORT:=8080}"
: "${DEBUG:=0}"

# ✅ Default value only if unset (keeps empty string)
: "${NAME:=default}"

# ✅ Use in assignments
config="${CONFIG_FILE:-/etc/default.conf}"

# ✅ Error if variable unset
: "${REQUIRED_VAR:?Variable REQUIRED_VAR must be set}"

# ✅ Alternate value if set
message="${DEBUG:+Debug mode enabled}"

# ✅ String operations
filename="document.txt"
basename="${filename%.*}"      # document
extension="${filename##*.}"    # txt
```

**Techniques:**

- Use `${VAR:-default}` for default if unset or empty
- Use `: "${VAR:=value}"` to set defaults
- Use `${VAR:?message}` to require variables
- Use `${VAR:+value}` for alternate values when set
- Use `${var%pattern}` and `${var##pattern}` for string operations
