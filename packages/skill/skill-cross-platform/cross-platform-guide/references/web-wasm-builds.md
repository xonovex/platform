# web-wasm-builds: Building Native C/C++ to WebAssembly

**Guideline:** Treat the browser as just another platform backend, but restructure the parts the web cannot host: hand the frame loop to the browser via a cooperative main-loop callback, make all file and network access asynchronous, map your explicit-API renderer down to GL ES / WebGL (or WebGPU) — typically by routing only the 2D/UI path — and account for a 32-bit, single-address-space, single-threaded sandbox with growable but fully-committed memory.

**Rationale:** A Wasm/Emscripten toolchain compiles your portable C/C++ unchanged, but the browser owns the event loop, the network is non-blocking, the filesystem is virtual, threads are restricted, and pointers are 32-bit — so the platform-specific delta is concentrated and large rather than diffuse. A native `while (running) { tick(); }` loop blocks the browser's event loop and hangs the tab; the browser must drive the frame. Synchronous `fopen`/`read` and blocking sockets do not exist over the network, so assets must be preloaded into the virtual filesystem or fetched asynchronously. The renderer is the deepest cut: browsers expose WebGL (OpenGL ES), not a low-level explicit API, so a Vulkan/D3D12/Metal engine cannot run its native backend — the pragmatic path is to map only the UI/2D draw stream to a tiny GL ES pipeline and defer or stream 3D. Getting these structural constraints right is what makes the same source that runs natively also run in a tab.

## Contents

- Toolchain and serving — output files, local server requirement
- Build flags and platform defines — selecting the web backend
- Cooperative main loop — handing the frame to the browser
- Async file and network access — preload vs fetch, virtual filesystem
- Memory model — 32-bit pointers, growth, committed memory
- Rendering — mapping an explicit-API renderer down to GL ES / WebGPU
- Input — browser event callbacks into the input abstraction
- Linking — static-only, dynamic-plugin and symbol-collision pitfalls

## Toolchain and Serving

The Wasm toolchain emits three artifacts: a `.wasm` binary, a `.js` loader/glue file, and an `.html` entry page. Browsers refuse to load Wasm from `file://` for security, so you must serve over HTTP even for local testing (any static server, e.g. `python3 -m http.server`, works). Program `stdout` can be redirected to a DOM element (e.g. a textarea), so command-line tools and the foundation library run and print in the browser before any graphics exist.

## Build Flags and Platform Defines

Add `web` as a real platform in the build system: give it its own defines, set the output extension to `.html`, and select the web OS backend translation unit. Because the browser sandbox is broadly POSIX-like under Emscripten, the web backend often starts as a copy of the POSIX/Linux backend (e.g. `os.linux.c` → `os.web.c`), then diverges only where the browser differs (loop, files, memory).

```text
# Conceptual flag set for a web build:
-D PLATFORM_WEB -D PLATFORM_POSIX -D NO_MAIN_FIBER   # platform selection + disable native main fiber
-s ALLOW_MEMORY_GROWTH=1                              # let the heap grow past the initial size
--preload-file assets@/data                           # pack ./assets into the virtual FS at /data
-o app.html                                           # emit .wasm + .js + .html
```

## Cooperative Main Loop

The browser, not your program, owns the frame clock. Register a per-frame callback and return control to the browser each frame instead of spinning a blocking `while` loop. Disable any native "main fiber" / blocking-loop mechanism for the web build (the platform define above), and move per-frame work into the callback.

```c
#include <emscripten.h>

static void frame(void) {
    pump_input();
    update(dt);
    render();          // returns control to the browser after each frame
}

int main(void) {
    init();
    // fps=0 -> use requestAnimationFrame; simulate_infinite_loop=1 -> main() does not "return" to teardown
    emscripten_set_main_loop(frame, 0, 1);
    return 0;          // unreachable while the loop runs
}
```

## Async File and Network Access

There is no synchronous blocking file/network read in the browser. Two strategies:

- **Preload into the virtual filesystem** at link time (`--preload-file dir@/mount`); the packaged files then appear under the mount point so ordinary `fopen("data/font.ttf", "rb")` works after the runtime initializes.
- **Fetch asynchronously** at runtime for data not known at build time; the request returns immediately and a callback delivers the bytes later. Restructure code that assumed a blocking load into request → continue → callback.

## Memory Model

Pointers are 32-bit, which changes `sizeof(struct)` wherever a pointer sits inside one — re-audit struct padding, hashing, and `memcmp` (see porting-strategy). The heap is a single contiguous buffer; enable `ALLOW_MEMORY_GROWTH=1` to let it expand past the initial size. Crucially, all requested memory is actually committed, so virtual-memory tricks that reserve a huge address range and commit lazily do not work — replace a reserve-then-commit allocator with a fixed, bounded allocation on the web build.

```c
#if defined(PLATFORM_WEB)
    // No lazy commit: cap the arena instead of reserving a huge VM range.
    allocator = create_fixed_arena(1 * 1024 * 1024, scope);
#else
    allocator = create_reserving_vm_arena(GIB(16), scope); // reserve large, commit on demand
#endif
```

## Rendering

Browsers expose WebGL (OpenGL ES), not an explicit low-level API, so a Vulkan/D3D12/Metal renderer cannot run its native backend in a tab (see gpu-rendering-guide / gpu-rendering-vulkan-guide for the native side). The practical path is to map only the UI/2D draw stream: convert the compact GPU-side vertex format (e.g. a rectangle stored as `{x, y, w, h}`) into "fat" expanded vertices on the CPU, feed them through a minimal GL ES vertex/fragment shader pair with clip-rect and texture sampling, and accept the higher memory/bandwidth cost as the price of correctness. Text and textured glyphs (e.g. via stb_truetype) ride the same expanded-vertex path. WebGPU is the forward-looking target that maps an explicit API more faithfully, but WebGL is the broadly-available floor today. Full 3D viewports are typically deferred or rendered server-side and streamed.

## Input

Wire the browser's event callbacks into the same input abstraction the native backends fill, synthesizing the engine's input events from DOM events.

```c
EM_BOOL on_mouse(int type, const EmscriptenMouseEvent *e, void *ud) {
    input_event_t ev = {.type = INPUT_DATA_CHANGE, .source = &mouse_source};
    // map EMSCRIPTEN_EVENT_MOUSEMOVE / MOUSEDOWN / MOUSEUP into input items
    submit_input_event(&ev);
    return EM_TRUE;
}
// emscripten_set_mousemove_callback(...) etc.
```

## Linking

Web builds are statically linked — there is no native dynamic-plugin loader by default. Code that normally loads subsystems as plugins now links them all into one binary, which surfaces duplicate symbols: every plugin defining `load_plugin()` or a global API pointer collides. The pragmatic fix is `#if defined(PLATFORM_WEB)` guards to drop the duplicates; the principled fix is to support the toolchain's `dlopen` so plugins load as on native. Some browser APIs are partial: callbacks that link on one browser may fail to link on another (e.g. touch callbacks), so test on each target browser, not just one.

**Gotchas:**

- A blocking `while (running)` loop hangs the tab; the browser owns the frame clock — you must yield via the main-loop callback every frame.
- 32-bit pointers change struct padding silently, breaking any hash/`memcmp` that assumed a 64-bit layout; widen keys and re-pad for the web build.
- Reserve-then-commit virtual-memory allocators do not work — Emscripten commits everything requested; switch to a bounded fixed allocator on web.
- `ALLOW_MEMORY_GROWTH` lets the heap grow but invalidates cached raw pointers into the heap on a growth event in some glue configurations; re-fetch base pointers rather than caching them across allocations.
- Assets are not on a real disk: a `fopen` that "works natively" returns nothing in the browser unless the file was preloaded into the virtual FS or fetched asynchronously first.
- Static-linking everything that was a plugin creates duplicate-symbol collisions (`load_plugin`, global API pointers); guard or properly `dlopen`, don't ignore the linker.
- Browser API coverage differs per browser; a callback that links in one may produce link errors in another — CI must build/run on each target browser.

**Related:** [references/platform-abstraction-layer.md](./platform-abstraction-layer.md), [references/porting-strategy.md](./porting-strategy.md), **gpu-rendering-guide**, **gpu-rendering-vulkan-guide**
