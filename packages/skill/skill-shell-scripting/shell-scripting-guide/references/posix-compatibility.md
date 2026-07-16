# posix-compatibility: POSIX Compatibility

For portable `sh`: use `tr`/`cut` instead of bash `${var^^}`/`${var,,}`. No arrays — use positional params or temp files. Test with `dash`, not just bash. Declare a bash dependency with `#!/usr/bin/env bash` rather than relying on bashisms under `sh`.

```sh
upper="$(printf '%s' "$string" | tr '[:lower:]' '[:upper:]')"   # ✅ POSIX, no ${var^^}
```
