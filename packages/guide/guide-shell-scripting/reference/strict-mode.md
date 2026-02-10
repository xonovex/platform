# strict-mode: Strict Mode and Safety

**Guideline:** Enable strict mode with `set -eu` (POSIX) or `set -euo pipefail` (bash/zsh) to catch errors early.

**Rationale:** Strict mode prevents scripts from continuing after errors, catches undefined variable usage, and makes pipeline failures visible, preventing silent failures that can cause data loss or security issues.

**Example:**

```sh
# ✅ Basic strict mode (POSIX)
set -eu

# ✅ Full strict mode (bash/zsh)
set -euo pipefail

# What each flag does:
# -e: Exit immediately if command fails
# -u: Error on undefined variable
# -o pipefail: Fail if any command in pipeline fails

# ✅ Safer alternative for POSIX (no pipefail)
set -eu
IFS=$(printf '\n\t')

# ✅ Handle expected failures explicitly
if ! command_that_might_fail; then
    log "Expected failure occurred"
fi

# ✅ Temporary disable strict mode
set +e
optional_command || true
set -e
```

**Techniques:**

- Add `set -eu` at the top of POSIX scripts
- Add `set -euo pipefail` for bash/zsh scripts
- Handle expected failures explicitly with `if ! command; then`
- Temporarily disable strict mode only when necessary
- Set IFS to newline and tab for POSIX compatibility
