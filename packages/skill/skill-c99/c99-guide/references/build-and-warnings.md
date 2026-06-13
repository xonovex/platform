# Build dialect & warnings

A clean C99 build pins the language, exposes the platform surface deliberately, and treats warnings as errors — without fighting the idioms C99 is built on.

## Pin strict ISO C99, not the GNU dialect

`-std=c99` alone is not strict: most toolchains default to the GNU dialect (`-std=gnu99`) and quietly accept extensions. Pin it:

```cmake
set_target_properties(target PROPERTIES
  C_STANDARD 99 C_STANDARD_REQUIRED ON C_EXTENSIONS OFF)   # -> -std=c99
```

## Strict C99 hides POSIX — re-expose it explicitly

Under strict `-std=c99` the libc headers hide everything outside ISO C: `readlink`, `strnlen`, `ssize_t`, `pthread_*`, `clock_gettime`, etc. Calling them then becomes an **implicit-declaration error**, not a warning. Re-expose the POSIX surface with a feature-test macro instead of falling back to the GNU dialect:

```c
#define _XOPEN_SOURCE 700   /* SUSv4 / POSIX.1-2008 + XSI; before any include */
```

Set it as a compile definition on the target (so it precedes every include) rather than per-file. `_XOPEN_SOURCE=700` implies `_POSIX_C_SOURCE=200809L` and covers the common backend needs.

## Warnings as errors — `-Wall -Wextra -Werror`

Turn the build strict, but suppress the warnings that fight idiomatic C99 rather than catch bugs:

```
-Wall -Wextra -Werror -Wno-missing-field-initializers -Wno-missing-braces
```

- **`-Wmissing-field-initializers` / `-Wmissing-braces`** (both from `-Wextra`) flag intentional zero-is-initialization — partial designated initializers and `= {0}` over nested aggregates. ZII is a core C99 idiom; suppress these two, keep the rest.
- **The unused-symbol family** (`-Wunused-parameter`, `-Wunused-variable`, `-Wunused-function`, `-Wunused-but-set-variable`) is mostly noise for **library** code, which legitimately carries surface a translation unit never references — interface-mandated callback parameters, reflection/mapping tables defined in headers (compiled into every includer), helpers built for callers. Relax these for library targets; keep them for leaf/application code, where an unused symbol is a real dead-code smell. Either way, keep **`-Wunused-value`** (a computed-then-discarded expression is a real bug).

`-Wall -Wextra` do **not** include `-Wsign-conversion` (nor `-Wconversion`). Enable `-Wsign-conversion` to catch the signed/unsigned mix-ups exact-width discipline cares about — chiefly comparing or assigning a signed `int` to a `size_t` count/index, where a negative value silently becomes huge. It can be noisy on existing code; add it deliberately and fix at the type level (keep counts/indices `size_t` end to end) rather than papering over it with casts.

## Detect `snprintf` truncation (and silence `-Wformat-truncation`)

`snprintf` into a fixed buffer is safe but may truncate. GCC's `-Wformat-truncation` (under `-Wall` at `-O1`+) flags it. Check the return value — this both handles the truncation correctly and satisfies the warning:

```c
int n = snprintf(dst, sizeof(dst), "%s/%s", dir, name);
if (n < 0 || (size_t)n >= sizeof(dst)) {
    return false;  /* result too long for dst */
}
```

A bare unchecked `snprintf` is the pattern the warning exists to catch.

## Sanitizers — keep a debug build under ASan + UBSan

Warnings catch what the compiler can prove statically; sanitizers catch the rest at runtime. Keep a debug variant compiled and linked with them, and run the test suite under it:

```cmake
target_compile_options(target_asan PRIVATE -fsanitize=address,undefined -fno-omit-frame-pointer -g)
target_link_options(target_asan    PRIVATE -fsanitize=address,undefined)
```

- **AddressSanitizer** (`-fsanitize=address`) red-zones heap, stack, and global allocations to catch overflows, use-after-free, and double-free — the off-by-one that _doesn't_ segfault. Add `ASAN_OPTIONS=detect_leaks=1` for leak detection.
- **UndefinedBehaviorSanitizer** (`-fsanitize=undefined`) traps signed overflow, misaligned access, out-of-range shifts, and invalid `enum`/`bool` values — the UB the optimizer would otherwise silently exploit.
- Keep `-fno-omit-frame-pointer -g` so reports carry real stack traces. ASan and ThreadSanitizer are mutually exclusive — give each its own build preset.
- This is the standing answer to "track a crash with `printf`": reproduce it under ASan or a debugger (`gdb`/`lldb`), which name the faulting line and the call stack directly. The `compile_commands.json` from `CMAKE_EXPORT_COMPILE_COMMANDS=ON` wires clangd and IDE debuggers to do this with no extra setup.

## One translation unit or many

A unity build (`#include`-ing every `.c` into one TU, or CMake's `UNITY_BUILD`) gives a single fast clean compile and whole-program inlining — convenient for a small program or a one-file amalgamation. The cost is incremental builds: touching any file recompiles everything, and TU-local `static` symbols leak into one namespace. For anything that outgrows a handful of files, a structured multi-file build with a disciplined header graph keeps incremental builds fast — prefer it, and treat unity as a deliberate small-scope choice, not the default.
