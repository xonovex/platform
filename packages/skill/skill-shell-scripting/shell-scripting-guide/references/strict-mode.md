# strict-mode: Strict Mode and Safety

Handle expected failures explicitly so `-e` doesn't abort; scope any `set +e` tightly.

```sh
set -euo pipefail

if ! command_that_might_fail; then      # expected failure — doesn't trip -e
    log "handled"
fi

set +e; optional_command; set -e        # temporarily relax, then restore
```
