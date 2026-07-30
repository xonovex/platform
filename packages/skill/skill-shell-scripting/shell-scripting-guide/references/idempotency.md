# idempotency: Idempotent Scripts

Make scripts safe to re-run: check-before-create so a second run is a no-op.

```sh
[ -d "$dir" ] || mkdir -p "$dir"
```
