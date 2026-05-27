# Sources

## Game-engine development blog (archive)

- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Essentials, Abstraction and Porting, Web / WebAssembly, Input Devices, Gotchas
  - A worked, shipped instance of porting a native C engine to Linux and the web, and adding gamepad input, all behind one OS-abstraction interface — generalized away from the engine to neutral portability guidance
- **Aspects extracted:**
  - "Porting the engine to Linux" — one OS-abstraction interface (general OS services, windowing, clipboard, input source, audio backend) so app code never `#ifdef`s; porting in interactivity order (window first, then input, then OS services, then audio/dialogs); emulating concepts the OS lacks (relative mouse motion, selection-based clipboard); window managers overriding size requests; leaning on portable libs for fibers/file dialogs; recording unfinished work → `references/platform-abstraction-layer.md`, `references/porting-strategy.md`
  - "Compiling the engine with Emscripten — Part 1" — adding `web` as a build platform with its own defines/output; copying the POSIX/Linux OS backend to a web backend; stricter compiler (unused vars, `void f(void)`); 32-bit pointer fallout on struct padding and on hash keys (the `void*`/`uint64_t` union widening); serving over HTTP, stdout to a DOM element → `references/web-wasm-builds.md`, `references/porting-strategy.md`
  - "Compiling the engine with Emscripten — Part 2" — cooperative main loop replacing the native blocking loop; preload-file virtual filesystem vs async fetch; mapping a Vulkan renderer's compact UI vertex stream down to expanded "fat" vertices through a minimal WebGL/GL ES shader pair; text via stb_truetype on that path; `ALLOW_MEMORY_GROWTH` and the failure of reserve-then-commit virtual-memory tricks (fixed bounded allocator instead); static-linking-only with duplicate-symbol collisions and the `dlopen` direction; per-browser API coverage (touch callback link errors); wiring Emscripten mouse callbacks into the input abstraction → `references/web-wasm-builds.md`
  - "Gamepad Implementation on Linux" — evdev enumeration under `/dev/input`, preferring `by-id`/`by-path` symlinks (`-event-joystick`); hotplug via filesystem-change monitoring; capability probing with `ioctl`/`EVIOCGBIT`/`BTN_GAMEPAD`; per-axis `EVIOCGABS` range and `flat` deadzone; non-blocking `select`/`poll` + `read` of `struct input_event`; the code-keyed mapping table expressing 1:1, 2:1, 1:2 button/axis mappings; routing through the input-source interface → `references/input-devices.md`, `references/platform-abstraction-layer.md`

## Emscripten / WebAssembly documentation

- **URL:** https://emscripten.org/docs/index.html
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Web / WebAssembly
  - Confirming the general toolchain behavior independent of any one engine: `emscripten_set_main_loop`, `--preload-file`, `-s ALLOW_MEMORY_GROWTH`, the `.wasm`/`.js`/`.html` output triple, the virtual filesystem, and async fetch
- **Aspects extracted:**
  - Cooperative main loop, virtual filesystem preloading vs runtime fetch, memory growth flag, output artifacts, WebGL/GL ES as the available graphics floor and WebGPU as the explicit-API direction → `references/web-wasm-builds.md`

## Linux evdev input subsystem documentation

- **URL:** https://www.kernel.org/doc/html/latest/input/input.html
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Input Devices
  - Confirming evdev semantics independent of any one engine: `/dev/input/eventN`, `struct input_event` (type/code/value), `ioctl` capability queries (`EVIOCGBIT`, `EVIOCGABS`, `struct input_absinfo` with `flat` deadzone), and `BTN_*`/`ABS_*` code namespaces
- **Aspects extracted:**
  - Device node model, event struct, capability/axis probing, deadzone metadata, button/axis code namespaces → `references/input-devices.md`

## Refresh Workflow

1. Re-read the upstream source(s) above (the four blog posts, Emscripten docs, the kernel input docs)
2. Diff against the prior pull (or scan for newly added sections / API revisions)
3. For each changed area, update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
