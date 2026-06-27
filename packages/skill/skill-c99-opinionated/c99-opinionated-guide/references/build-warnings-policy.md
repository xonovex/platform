# Build & warnings policy

This style ships `-Wall -Wextra -Werror`, but a library's warning policy differs from an application's, and strict C99 needs help to see the platform.

## Strict ISO C99 + explicit POSIX

Pin the standard so the toolchain can't drift into the GNU dialect, then re-expose POSIX deliberately:

```cmake
set_target_properties(target PROPERTIES C_STANDARD 99 C_STANDARD_REQUIRED ON C_EXTENSIONS OFF)
target_compile_definitions(target PRIVATE _XOPEN_SOURCE=700)  # readlink/strnlen/ssize_t/pthread/clock_gettime
```

Without `_XOPEN_SOURCE` (or equivalent), strict `-std=c99` turns every POSIX call into an implicit-declaration error. Centralize this in one `strict_c(target)` helper and apply it to every first-party target so the policy is one edit, not N.

## Unused symbols are library surface, not defects

The decisive split: **`-Werror` for correctness warnings, but the unused-_symbol_ family is not an error for library code.**

```
-Werror -Wno-unused-parameter -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function
        -Wno-missing-field-initializers -Wno-missing-braces   # ZII is intentional
```

A library legitimately carries surface its own translation units never reference:

- interface-mandated callback/vtable parameters,
- reflection / enum-name / backend-mapping tables defined in **headers** — these compile into _every_ includer, so a consumer that doesn't use one would otherwise fail `-Werror` on the library's header,
- `_DEFAULT` consts and helpers built for callers.

Forcing `(void)param;` and `__attribute__((unused))` across that surface is noise that fights the nature of a library. So relax the unused-symbol family for it — but keep **`-Wunused-value`** (a discarded computation is a real bug) and every correctness warning as a hard error. Leaf/application targets keep the full set: there, an unused symbol _is_ dead code.

This is the one place the "fix every warning" reflex is wrong: for a library, unused surface is the point.

## snprintf truncation

`snprintf` is bounded but may truncate; check the return both to handle it and to clear `-Wformat-truncation`:

```c
int n = snprintf(dst, sizeof(dst), "%s/%s", a, b);
if (n < 0 || (size_t)n >= sizeof(dst)) return ERR_PATH_TOO_LONG;
```

## Sanitizers — a debug preset under ASan + UBSan

Strict warnings catch what the compiler proves statically; manual memory and pointer/index work need a runtime net too. Keep a debug preset built with the sanitizers and run the tests under it:

```cmake
target_compile_options(target_asan PRIVATE -fsanitize=address,undefined -fno-omit-frame-pointer -g)
target_link_options(target_asan    PRIVATE -fsanitize=address,undefined)
```

- **AddressSanitizer** red-zones allocations and stack frames to catch the overflow / use-after-free / double-free that doesn't segfault — exactly the failure mode of hand-carved arena and caller-owned buffers. `ASAN_OPTIONS=detect_leaks=1` adds leak detection.
- **UBSan** traps signed overflow, misaligned access, and out-of-range shifts — relevant when you do alignment math or pointer tagging.
- ASan and **ThreadSanitizer** are mutually exclusive; give each its own preset. Keep `-fno-omit-frame-pointer -g` for readable traces, and reproduce crashes under ASan or a debugger rather than `printf`.

## Caller-owns extends to strings

The `_req()`/`_init()` sizing pattern (see caller-owns-memory) covers string building too: a builder takes `_req(max_len)` bytes the caller allocates, then every append is bounded against that capacity and latches a `truncated` flag — no hidden allocation, no `strcat` overrun. Reads borrow length-carrying views; only the boundary does the one `strlen`. Full treatment in [references/string-handling.md](./string-handling.md).

### Related

[references/caller-owns-memory.md](./caller-owns-memory.md), [references/string-handling.md](./string-handling.md), [references/safety-validations.md](./safety-validations.md)
