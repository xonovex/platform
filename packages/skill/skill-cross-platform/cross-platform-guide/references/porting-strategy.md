# porting-strategy: Order of Operations for a New Target

**Guideline:** Port to a new OS by working through the abstraction interface in interactivity order — get a window on screen, then input, then the rest of the OS services, then audio/dialogs — stubbing each interface first and replacing stubs incrementally, while pinning down toolchain differences (compiler strictness, pointer width, struct layout) and wiring a per-target CI build before declaring the platform done.

**Rationale:** Porting is mostly discovery, and discovery is fastest when you can see and poke the program: a visible, interactive window turns every subsequent gap into an observable symptom instead of a link error in the dark, which is why "get something on screen" beats "implement everything alphabetically." Stub-first keeps the whole program compiling and linking from hour one, so each newly implemented backend function is a small, testable delta rather than a big-bang integration. The other half of porting is the toolchain: a different compiler (GCC/Clang vs MSVC, or Emscripten's stricter Clang) flags real latent bugs and rejects sloppy declarations, and a different pointer width or struct-packing assumption breaks code that hashed or `memcmp`-ed structs assuming a fixed layout. Catching these per target — and re-catching them automatically in CI for every target — is what stops a port from silently rotting the moment someone edits portable code on a different host.

**How to Apply:**

1. Add the new platform to the build system as a first-class config (its own defines, output type, and compiler invocation); start it linking against the default stub backend so it runs immediately.
2. Implement the window/graphics backend first to get a visible surface, then the input source, then the remaining general OS services (threading, fibers, time, filesystem, file-change monitoring), then audio and native file dialogs.
3. Treat each compiler warning as a real defect: fix unused variables, write `void f(void)` not `void f()` in C, and resolve declaration mismatches the new compiler is stricter about rather than blanket-suppressing.
4. Audit anything that assumes a pointer width or struct layout — hashing, `memcmp`, explicit padding macros, serialized blobs — because a 32-bit target (e.g. Wasm) or different ABI changes padding and key sizes.
5. Lean on existing portable libraries for fiddly subsystems (fibers, file dialogs) instead of hand-rolling per platform, isolating them inside the backend.
6. Add the new target to CI as its own matrix entry so every later commit is built (and ideally smoke-tested) for it; a port that is not in CI regresses silently.
7. Record what is intentionally unfinished (joysticks, multi-mouse, an untested window manager / compositor) so missing features read as known gaps, not bugs.

**Example:**

```c
// Pointer-width / layout pitfalls a new toolchain exposes.
// 1) A 32-bit target changes struct padding that code relied on for hashing/memcmp.
struct key { void *ptr; uint32_t tag; };   // 16 bytes on LP64, 8 bytes on a 32-bit (Wasm) target

// 2) A hash table keyed on 64-bit values must widen a 32-bit pointer explicitly:
typedef union { void *ptr; uint64_t u64; } ptr_key_t;
uint64_t k = ((ptr_key_t){.ptr = p}).u64;  // portable across 32/64-bit pointer widths

// 3) C declaration the stricter compiler rejects:
//   void shutdown();      // K&R: "unspecified args" — warns/errors on a strict toolchain
void shutdown(void);       // correct: explicitly no parameters
```

**Gotchas:**

- "It compiles on the old platform" says nothing about the new one; a stricter compiler surfacing fresh warnings usually means the warnings were always real, not that the new compiler is wrong.
- Suppressing a warning per target (e.g. disabling the padding warning for one build) unblocks the port but leaves a latent layout assumption — record it as debt, do not call it fixed.
- A 32-bit target silently changes `sizeof(struct)` wherever a pointer sits inside it, breaking anything that hashed or compared the bytes; widen keys and re-pad deliberately.
- Implementing OS services in alphabetical/struct order instead of interactivity order leaves you debugging blind for days before anything is visible.
- A new platform that is not in CI builds green locally and rots on the next portable-code edit; add the matrix entry before moving on.
- Window managers, desktop environments, and compositors are free to override window size/position requests; never assume the size you asked for is the size you got — query it back.

**Related:** [references/platform-abstraction-layer.md](./platform-abstraction-layer.md), [references/web-wasm-builds.md](./web-wasm-builds.md), **cmake-guide**, **c99-opinionated-guide**
