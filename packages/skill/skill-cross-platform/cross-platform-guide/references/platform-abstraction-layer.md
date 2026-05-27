# platform-abstraction-layer: A Single Interface for All OS Calls

**Guideline:** Route every operating-system and windowing call through one narrow abstraction interface — a struct of function pointers (or a header whose implementation is selected at build time) — and forbid application, renderer, and engine code from ever calling an OS API directly or guarding logic with platform `#ifdef`s.

**Rationale:** Conditional compilation scattered through application code is the central failure of ad-hoc ports: every feature ends up knowing every platform, so each new target multiplies the surface that must be re-read and re-tested, and platform bugs hide inside otherwise-portable logic. Concentrating OS access behind one interface inverts this — application code depends only on the abstract contract, and a port becomes "write one new backend that fills the interface," with zero edits to callers. Because the interface is the unit of completeness, you can stub it, compile against the stub on day one, and replace stubs one function at a time while the rest of the system keeps building. Grouping by concern (windowing, clipboard, threading/fibers, file watching, time, input, audio) also keeps each backend small enough that a missing or fragile platform feature is isolated to its own struct, not smeared across the codebase.

**How to Apply:**

1. Carve the OS surface into cohesive interfaces by concern: one for general OS services (threading, fibers, time, filesystem, file-system change monitoring), one for windowing, one for clipboard/selection, one for input sources, one for the audio backend.
2. Express each as a struct of function pointers (so the same header compiles on every target) or as a header with a build-time-selected `.c`; expose only the abstract types callers need, never raw OS handles.
3. Have application/engine/renderer code take the interface by pointer and call through it — never call an OS API or include an OS header outside a backend file.
4. Provide a default stub backend that satisfies the interface with no-ops / sane defaults so the whole program links and runs before any real platform code exists.
5. Implement one backend per platform in its own translation unit (e.g. `os.linux.c`, `os.win32.c`, `os.web.c`); the build picks exactly one.
6. When the OS lacks a concept the interface assumes (e.g. relative mouse motion, a single system clipboard), emulate it inside the backend so the contract still holds for callers.

**Example:**

```c
// os.h — one abstract interface, identical on every platform.
struct os_api {
    // threading / fibers / time
    os_thread_t (*create_thread)(os_thread_fn *fn, void *ud);
    uint64_t (*now)(void);
    // filesystem + change monitoring (used for hot-reload and device hotplug)
    bool (*directory_entries)(const char *path, os_entry_t **out, uint32_t *n);
    os_watch_t (*detailed_changes)(const char *path);
};

struct os_window_api {
    os_window_t (*create)(const os_window_desc_t *desc);
    void (*set_title)(os_window_t w, const char *title);
};

// Callers depend only on the struct, never on a platform header or an #ifdef:
extern struct os_api *os_api;
uint64_t t = os_api->now();          // resolves to os.linux.c / os.win32.c / os.web.c

// The build selects exactly one backend translation unit; the application
// code above is byte-for-byte identical on every target.
```

**Gotchas:**

- A backend that leaks a raw OS handle (HWND, file descriptor, X connection) through the interface re-couples callers to one platform — keep the type opaque and convert at the boundary.
- The interface must be designed before the first backend, not reverse-engineered from one platform, or its shape silently bakes in that platform's assumptions (e.g. assuming one global clipboard, or that the OS reports relative mouse motion).
- Stubs that return plausible-but-wrong values (success when unimplemented) hide missing functionality; return an explicit "unimplemented" signal so gaps surface during bring-up.
- Splitting the surface too finely produces dozens of one-function interfaces and ceremony; too coarsely produces a god-interface no backend can fill incrementally — group by the subsystem that ships together.
- Some platforms have no native concept the interface names (relative mouse delta, multiple selections vs one clipboard); the backend must emulate it (e.g. track last mouse position to synthesize deltas) rather than push the gap up to callers.

**Related:** [references/porting-strategy.md](./porting-strategy.md), [references/input-devices.md](./input-devices.md), [references/web-wasm-builds.md](./web-wasm-builds.md), **c99-opinionated-guide**, **gpu-rendering-guide**
