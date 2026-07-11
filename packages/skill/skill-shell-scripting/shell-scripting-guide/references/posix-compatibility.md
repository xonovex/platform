# posix-compatibility: POSIX Compatibility

For portable `sh`: `[ ]` not `[[ ]]`, `=` not `==`, `-gt`/`-lt` for numbers, `command -v` not `which`/`type`, `tr`/`cut` instead of bash `${var^^}`/`${var,,}`. No arrays — use positional params or temp files. Test with `dash`, not just bash. Declare a bash dependency with `#!/usr/bin/env bash` rather than relying on bashisms under `sh`.

```sh
if [ "$status" = "success" ] && [ "$count" -gt 10 ]; then      # ✅ POSIX
    upper="$(printf '%s' "$string" | tr '[:lower:]' '[:upper:]')"
fi
if [[ $status == success ]]; then ... fi                        # ❌ bashism under sh
```
