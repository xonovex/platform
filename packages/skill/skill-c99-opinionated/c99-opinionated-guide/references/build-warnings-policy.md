# Build & warnings policy

An opinionated overlay on **c99-guide**'s build-and-warnings policy. Apply **c99-guide** for the shared foundation — pinning strict ISO C99 with `C_EXTENSIONS OFF`, re-exposing POSIX through `_XOPEN_SOURCE=700`, suppressing the intentional-ZII `-Wmissing-field-initializers` / `-Wmissing-braces` pair, and keeping an ASan/UBSan debug preset. This overlay adds only the house-style deltas: the strict-C99 helper, the library-vs-application warning split, and the snprintf-truncation check.

## Centralize strict C99 in one helper

Wrap **c99-guide**'s strict-C99 + explicit-POSIX setup in a single `strict_c(target)` helper and apply it to every first-party target, so the policy is one edit, not N.

## Unused symbols are library surface, not defects

A library's warning policy differs from an application's. For library code, `-Werror` on correctness warnings, but relax the whole unused-_symbol_ family (a library legitimately carries interface-mandated parameters, header-defined reflection/mapping tables, and `_DEFAULT` helpers its own TUs never reference). Keep **`-Wunused-value`** — a discarded computation is a real bug — as a hard error:

```
-Werror -Wno-unused-parameter -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function
        -Wno-missing-field-initializers -Wno-missing-braces   # ZII is intentional
```

Leaf/application targets keep the full set: there, an unused symbol _is_ dead code.

## snprintf truncation

`snprintf` is bounded but may truncate; check the return both to handle it and to clear `-Wformat-truncation`:

```c
int n = snprintf(dst, sizeof(dst), "%s/%s", a, b);
if (n < 0 || (size_t)n >= sizeof(dst)) return ERR_PATH_TOO_LONG;
```

### Related

**c99-guide**, **memory-management-guide**, [safety-validations.md](safety-validations.md)
