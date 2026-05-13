# validation: Shellcheck and Shfmt Integration

**Guideline:** Validate scripts with shellcheck and format with shfmt for consistent, error-free code.

**Rationale:** Shellcheck catches common bugs, portability issues, and anti-patterns automatically. Shfmt ensures consistent formatting across scripts, improving readability and maintainability.

**Example:**

```sh
# Run shellcheck on script
shellcheck script.sh

# Check all shell scripts in directory
find . -name '*.sh' -exec shellcheck {} +

# Disable specific warnings (use sparingly)
# shellcheck disable=SC2086
variable_without_quotes=$1

# Better: Fix the actual issue
variable_with_quotes="$1"
```

**Techniques:**

- Run `shellcheck script.sh` before committing
- Use `find . -name '*.sh' -exec shellcheck {} +` to check multiple scripts
- Fix shellcheck warnings rather than disabling them
- Format scripts with `shfmt -w script.sh`
- Use shfmt options: `-i 2` (2-space indent), `-bn` (binary ops), `-ci` (case indent), `-sr` (space after redirect)
- Integrate shellcheck and shfmt into CI/CD pipelines
