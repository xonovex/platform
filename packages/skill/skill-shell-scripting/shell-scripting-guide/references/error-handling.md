# error-handling: Error Handling

Exit with a meaningful code: 0 success, 1 general, 2 usage/validation, 127 command not found. Validate file readability/existence before use.

```sh
die() { printf 'ERROR: %s\n' "$1" >&2; exit "${2:-1}"; }
[ -r "$file" ] || die "cannot read: $file" 1
```
