# input-devices: Enumerating and Reading Input Devices

**Guideline:** Treat device input as another platform backend behind the input-source interface: enumerate devices via the OS's device directory (on Linux, evdev nodes under `/dev/input`, preferring stable `by-id`/`by-path` symlinks), probe capabilities before trusting a device, read events non-blocking, detect hotplug by watching the device directory, and translate raw OS codes to your engine's buttons/axes through a data-driven mapping table rather than scattered conditionals.

**Rationale:** Input devices are heterogeneous, hot-pluggable, and OS-specific, so doing it inline couples gameplay/UI code to one platform's device API and to one controller's quirks. Routing through the same input-source interface the windowing backend uses means callers consume uniform events regardless of platform or device. Probing capabilities (which buttons and axes a device actually reports, and each axis's range and deadzone) before mapping prevents reading garbage from a device that is not the gamepad you assumed. Non-blocking reads keep the device poll inside the frame without stalling it. A mapping table keyed by the raw OS code is what tames the messy reality that one physical control can be 1:1 (a button), 2:1 (two axes → one stick), or 1:2 (one hat axis → two D-pad directions) — a table expresses all three uniformly, whereas `if/else` chains rot.

**How to Apply:**

1. Enumerate already-connected devices by listing the OS device directory; on Linux read `/dev/input` and prefer the descriptive `by-id`/`by-path` symlinks (joysticks/gamepads commonly end in `-event-joystick`) over raw `eventN` numbers, which are unstable across reboots/replugs.
2. Probe each candidate's capabilities before using it: query supported buttons and axes, and reject devices that are not actually the class you want (e.g. require `BTN_GAMEPAD` to confirm a gamepad).
3. For each axis, read its range and deadzone metadata (min, max, flat) so you can normalize raw values and apply the device-reported deadzone instead of a hard-coded guess.
4. Open device file descriptors non-blocking and poll them with `select()`/`poll()` plus `read()` each frame, draining all pending events; never block the frame on a device read.
5. Detect hotplug by monitoring the device directory for change events (the same filesystem-watch service used for hot-reload) and add/remove devices live.
6. Map raw OS codes to engine input items through a table keyed by the OS code, with entries that can express 1:1, 2:1 (component index into a vector), and 1:2 (a value splitting into a positive and a negative item) mappings.
7. Emit uniform engine input events through the input-source interface so callers never see evdev, `ioctl`, or any platform detail.

**Example:**

```c
// evdev: read range/deadzone per axis, then poll non-blocking, then map by code.
struct input_absinfo abs = {0};
if (ioctl(fd, EVIOCGABS(code), &abs) != -1) {
    axis.min = (float)abs.minimum;
    axis.max = (float)abs.maximum;
    axis.deadzone = (float)abs.flat;      // device-reported deadzone, not a guess
}

// Non-blocking poll: drain whatever is ready this frame, never block.
fd_set fds; FD_ZERO(&fds); FD_SET(fd, &fds);
struct timeval t = {0, 0};
if (select(FD_SETSIZE, &fds, NULL, NULL, &t) > 0) {
    struct input_event ev[64];
    int n = read(fd, ev, sizeof ev);      // type/code/value triples
    // ... feed ev[0..n) into the mapping below
}

// Data-driven mapping handles 1:1, 2:1 and 1:2 uniformly, keyed by the evdev code.
typedef struct axis_map_t {
    uint8_t positive_item;   // item set when value > 0 (or the only item, 1:1)
    uint8_t negative_item;   // item set when value < 0 (1:2 split, e.g. hat -> dpad L/R)
    uint8_t component_index; // which component of a vector item (2:1, e.g. X/Y -> left stick)
} axis_map_t;
```

**Gotchas:**

- Raw `eventN` numbers are reassigned across reboots and replugs; key persistent bindings off the stable `by-id`/`by-path` symlink, not the numeric node.
- A device node existing does not mean it is the device you want — a keyboard, mouse, and gamepad all appear under `/dev/input`; confirm the capability bits (e.g. `BTN_GAMEPAD`) before treating it as a gamepad.
- Hard-coding a deadzone ignores the per-device `flat` value the kernel already reports; read it and normalize per axis or sticks will drift or feel dead.
- A single physical control is not always one engine item: a hat axis splits 1:2 into two directions and X/Y axes fold 2:1 into one stick vector — only a mapping table expresses all cases without conditional sprawl.
- Blocking `read()` on a device fd stalls the frame; open non-blocking and gate the read on `select`/`poll`.
- Missing hotplug means controllers plugged in after launch never appear; watch the device directory for additions and removals, do not enumerate once at startup.
- Axis raw ranges differ per device and are often not symmetric; normalize using the queried min/max, never assume `[-32768, 32767]`.

**Related:** [references/platform-abstraction-layer.md](./platform-abstraction-layer.md), [references/porting-strategy.md](./porting-strategy.md), [references/web-wasm-builds.md](./web-wasm-builds.md)
