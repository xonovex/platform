# quoting: Quoting Variables and Expansions

Double-quote every variable expansion and command substitution to block word splitting and glob expansion. Quote `"$@"` in loops to preserve arguments containing spaces.

```sh
cat "$file"                    # ✅  cat $file breaks on "my file.txt"
for f in "$@"; do process "$f"; done   # ✅  unquoted $@ splits on spaces
current_date="$(date +%Y-%m-%d)"       # ✅ quote command substitution too
```
