---
name: shell-scripting-guide
description: "Use when writing or editing POSIX shell or Bash automation. Triggers on `.sh` files, `#!/usr/bin/env bash` shebangs, and prompts about scripting CI tasks, build helpers, devops automation, or one-off pipelines, even when the user doesn't say 'shell'. Apply for strict mode, quoting, parameter expansion, exit codes, idempotency, shellcheck/shfmt. Skip fish/zsh-only features and PowerShell."
---

# Shell Scripting Coding Guidelines

## Essentials

- **POSIX compatibility** - Use POSIX sh, lint with shellcheck, format with shfmt, see [references/posix-compatibility.md](references/posix-compatibility.md)
- **Strict mode** - Use strict mode and safe defaults, see [references/strict-mode.md](references/strict-mode.md)
- **Quoting** - Quote all expansions to avoid word splitting, see [references/quoting.md](references/quoting.md)
- **Functions** - Write small, focused functions, see [references/functions.md](references/functions.md)
- **Parameter expansion** - Set default values and manipulate variables, see [references/parameter-expansion.md](references/parameter-expansion.md)
- **Error handling** - Implement exit codes and error messages, see [references/error-handling.md](references/error-handling.md)
- **Idempotency** - Make scripts safely re-runnable, see [references/idempotency.md](references/idempotency.md)

## Gotchas

- Unquoted variables word-split and glob-expand — `cp $file dest/` silently breaks on filenames with spaces; quote everything
- `set -e` doesn't catch errors inside `if`, `&&`, `||`, or pipelines without `-o pipefail` — combine `set -euo pipefail` at the top
- `[[ ]]` is bash/ksh-specific; POSIX `sh` needs `[ ]` with different quoting rules — `#!/usr/bin/env bash` makes the dep explicit
- `command -v` is portable for checking command existence; `which` varies across platforms and exits non-zero on missing
- `trap ... EXIT` runs on normal exit AND on errors with `set -e` — use it for cleanup; place it BEFORE the code that needs cleanup

## Progressive disclosure

- Read [references/posix-compatibility.md](references/posix-compatibility.md) - When ensuring portability across shells
- Read [references/strict-mode.md](references/strict-mode.md) - When setting up error handling and safety flags
- Read [references/quoting.md](references/quoting.md) - When variables expand incorrectly or word splitting occurs
- Read [references/functions.md](references/functions.md) - When organizing script logic or creating reusable code
- Read [references/parameter-expansion.md](references/parameter-expansion.md) - When setting default values or manipulating variables
- Read [references/error-handling.md](references/error-handling.md) - When implementing exit codes or error messages
- Read [references/argument-parsing.md](references/argument-parsing.md) - When parsing command-line flags or arguments
- Read [references/validation.md](references/validation.md) - When checking preconditions or input validity
- Read [references/idempotency.md](references/idempotency.md) - When scripts should be safely re-runnable
- Read [references/common-patterns.md](references/common-patterns.md) - When learning common shell idioms
- Read [references/script-template.md](references/script-template.md) - When starting a new shell script
