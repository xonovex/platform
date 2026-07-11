# idempotency: Idempotent Scripts

Make scripts safe to re-run: check-before-create (`mkdir -p`, guard with `[ -d ]`), skip a mutation when its effect is already present (`grep -qF` before appending a line), and write-then-`mv` for atomic file replacement so a crash never leaves a half-written file.

```sh
[ -d "$dir" ] || mkdir -p "$dir"

grep -qF "$line" "$file" 2>/dev/null || printf '%s\n' "$line" >> "$file"

temp="$(mktemp)"; generate_config > "$temp"; mv "$temp" "$config"   # atomic
```
