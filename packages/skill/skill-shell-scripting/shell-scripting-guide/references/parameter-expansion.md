# parameter-expansion: Default Values and Parameter Expansion

Use expansion instead of external commands: `${VAR:-default}` (value if unset/empty), `: "${VAR:=value}"` (assign default in place), `${VAR:?message}` (abort if unset), `${VAR:+alt}` (alt when set). Strip with `${var%pattern}` (shortest trailing) / `${var%%...}`, `${var#pattern}` (shortest leading) / `${var##...}`.

```sh
: "${PORT:=8080}"              # set default if unset
config="${CONFIG_FILE:-/etc/default.conf}"
: "${REQUIRED_VAR:?must be set}"
filename="document.txt"
name="${filename%.*}"         # document
ext="${filename##*.}"         # txt
```
