# audio-callback-thread: The Real-Time Audio Callback Contract

**Guideline:** Treat the OS audio callback (or the dedicated render thread that feeds it) as a hard real-time context: it must produce a fixed block of samples within a fixed deadline, and must never block, lock, allocate, free, page-fault, log, or make a syscall on its path — do bounded, predictable work only.

**Rationale:** The audio device drains its buffer at a constant rate (e.g. 44.1 kHz x channels). If the next block is not ready when the hardware needs it, the device replays stale samples or silence and the listener hears a click, pop, or dropout. The callback runs on a high-priority thread the OS schedules just ahead of the device's consumption; any unbounded operation — a lock the game thread holds, a `malloc` that hits the kernel, a page fault, a file read — can stall it past the deadline. Unlike a frame drop in rendering, an audio underrun is immediately and harshly audible. The discipline is identical to an interrupt handler: known-bounded work, no waiting on anything another thread might own indefinitely.

**How to Apply:**

1. Render in fixed-size chunks (a "render quantum") sized to comfortably fit inside the buffer the device drains; everything the callback touches must already be allocated and resident.
2. Pre-touch and pin all buffers and voice state at startup so the first callback never page-faults; never grow a data structure from inside the callback.
3. Forbid, on the audio path: `malloc`/`free`, mutexes the game thread can hold, file/network/console I/O, and anything that can sleep unboundedly. Receive all work via a lock-free queue (see command-handoff).
4. Keep the render loop's worst case bounded: a fixed voice-pool size caps the inner loop, so the per-block cost has a hard ceiling regardless of how many sounds the game requests.
5. If running a dedicated render thread rather than rendering inside the device callback, give it elevated priority and have it sleep on the device's "buffer low" event, waking only to top up the queue.

**Example:**

```c
// Dedicated render thread: top up the device's queue, then sleep on its event.
// No locks, no allocation, bounded work per iteration.
#define RENDER_QUANTUM (257 * 8) // ~46 ms @ 44.1 kHz; *8 for SIMD, prime-ish to dodge cache aliasing

static void render_thread_main(audio_backend_t *backend, mixer_t *mixer) {
  for (;;) {
    // Keep ~1.5 quanta queued ahead; if we have enough, wait on the device event.
    if (backend->remaining_samples(backend) > 3 * RENDER_QUANTUM / 2) {
      backend->wait(backend, 0.1f); // sleeps until buffer-low or timeout
      continue;
    }
    // render() reads only preallocated voice state + drains the command queue.
    float *block = mixer_render(mixer, RENDER_QUANTUM);
    backend->feed(backend, block, RENDER_QUANTUM); // hand finished samples to the device
  }
}
```

**Gotchas:**

- A lock is poison even if it is "almost never" contended: the one time the game thread holds it during the callback's deadline, you get an audible dropout — priority inversion makes this worse, not rarer.
- `malloc`/`free` are not real-time safe; a single allocation that trips the allocator's slow path or the kernel can blow the deadline. Allocate the voice pool and all scratch buffers once, up front.
- The first touch of a fresh page faults; lazily-allocated buffers underrun on their first use even though "nothing changed." Write to every buffer once at init.
- Sizing the queue-ahead target too tight invites starvation from ordinary OS scheduling jitter and high-latency devices (USB/Bluetooth audio); too loose adds latency. Tune it against the worst real device, not the dev machine.
- Don't `printf` or assert from the callback to "debug" a glitch — the I/O itself causes the glitch. Record diagnostics into a preallocated ring and inspect them off-thread.

**Related:** [references/command-handoff.md](./command-handoff.md), [references/mixing-and-buffers.md](./mixing-and-buffers.md), **lock-free-guide**, **memory-management-guide**
