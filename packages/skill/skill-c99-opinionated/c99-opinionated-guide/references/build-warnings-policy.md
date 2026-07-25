# Build & warnings policy

An overlay on **c99-guide**'s build-and-warnings reference, which owns the shared foundation: pinning strict ISO C99, re-exposing POSIX with `_XOPEN_SOURCE`, the `-Wall -Wextra -Werror` baseline, checking `snprintf` for truncation, and the ASan/UBSan debug preset. Follow that first; this file carries only the decisions this style adds on top.

## Centralize the strict-C99 setup in one helper

Wrap the standard pinning and the POSIX feature-test macro in a single `strict_c(target)` helper and apply it to every first-party target, so the policy is one edit rather than N. Getting it wrong is not a warning — under strict `-std=c99` every POSIX call becomes an implicit-declaration error — so it is worth making impossible to forget rather than repeating per target.

## Unused symbols are library surface, not defects

The decisive split: **`-Werror` for correctness warnings, but the unused-_symbol_ family is not an error for library code.**

```
-Werror -Wno-unused-parameter -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function
        -Wno-missing-field-initializers -Wno-missing-braces   # ZII is intentional
```

Forcing `(void)param;` and `__attribute__((unused))` across a library's interface surface is noise that fights the nature of a library. Keep **`-Wunused-value`** and every correctness warning as a hard error. Leaf and application targets keep the full set — there an unused symbol _is_ dead code.

This is the one place the "fix every warning" reflex is wrong: for a library, unused surface is the point.

## Caller-owns extends to strings

The `_req()`/`_init()` sizing pattern covers string building too: a builder takes `_req(max_len)` bytes the caller allocates, then every append is bounded against that capacity and latches a `truncated` flag — no hidden allocation, no `strcat` overrun. Reads borrow length-carrying views; only the boundary does the one `strlen`. Full treatment in [references/string-handling.md](./string-handling.md).

### Related

[references/caller-owns-memory.md](./caller-owns-memory.md), [references/string-handling.md](./string-handling.md), [references/safety-validations.md](./safety-validations.md)
