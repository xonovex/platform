---
name: shell-scripting-guidelines
description: >-
  Trigger on `.sh` files and shell scripts. Use when writing POSIX shell scripts for automation. Apply for strict mode, quoting, error handling, idempotent operations. Keywords: shell, bash, POSIX, shellcheck, shfmt, strict mode, quoting, parameter expansion, exit codes, idempotency.
---

# Shell Scripting Coding Guidelines

## Essentials

- **POSIX compatibility** - Use POSIX sh, lint with shellcheck, format with shfmt, see [reference/posix-compatibility.md](reference/posix-compatibility.md)
- **Strict mode** - Use strict mode and safe defaults, see [reference/strict-mode.md](reference/strict-mode.md)
- **Quoting** - Quote all expansions to avoid word splitting, see [reference/quoting.md](reference/quoting.md)
- **Functions** - Write small, focused functions, see [reference/functions.md](reference/functions.md)
- **Parameter expansion** - Set default values and manipulate variables, see [reference/parameter-expansion.md](reference/parameter-expansion.md)
- **Error handling** - Implement exit codes and error messages, see [reference/error-handling.md](reference/error-handling.md)
- **Idempotency** - Make scripts safely re-runnable, see [reference/idempotency.md](reference/idempotency.md)

## Progressive disclosure

- Read [reference/posix-compatibility.md](reference/posix-compatibility.md) - When ensuring portability across shells
- Read [reference/strict-mode.md](reference/strict-mode.md) - When setting up error handling and safety flags
- Read [reference/quoting.md](reference/quoting.md) - When variables expand incorrectly or word splitting occurs
- Read [reference/functions.md](reference/functions.md) - When organizing script logic or creating reusable code
- Read [reference/parameter-expansion.md](reference/parameter-expansion.md) - When setting default values or manipulating variables
- Read [reference/error-handling.md](reference/error-handling.md) - When implementing exit codes or error messages
- Read [reference/argument-parsing.md](reference/argument-parsing.md) - When parsing command-line flags or arguments
- Read [reference/validation.md](reference/validation.md) - When checking preconditions or input validity
- Read [reference/idempotency.md](reference/idempotency.md) - When scripts should be safely re-runnable
- Read [reference/common-patterns.md](reference/common-patterns.md) - When learning common shell idioms
- Read [reference/script-template.md](reference/script-template.md) - When starting a new shell script
