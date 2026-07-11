# strict-mode: Strict Mode and Safety

Start with `set -eu` (POSIX) or `set -euo pipefail` (bash/zsh). `-e` exits on failure, `-u` errors on unset variable, `-o pipefail` propagates a mid-pipeline failure (not POSIX — bash/zsh only). Handle expected failures explicitly so `-e` doesn't abort; scope any `set +e` tightly.

```sh
set -euo pipefail

if ! command_that_might_fail; then      # expected failure — doesn't trip -e
    log "handled"
fi

set +e; optional_command; set -e        # temporarily relax, then restore
```
