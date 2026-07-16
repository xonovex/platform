# audio-callback-thread: The Real-Time Audio Callback Contract

## Guideline

Treat the OS audio callback (or the dedicated render thread that feeds it) as a hard real-time context: it must produce a fixed block of samples within a fixed deadline, and everything it touches must already be allocated and resident (pre-touched) so it never page-faults on its path.

## Rationale

The device drains its buffer at a constant rate; if the next block isn't ready it replays stale samples or silence, and an audio underrun is immediately and harshly audible — far worse than a dropped render frame. The callback runs on a high-priority thread scheduled just ahead of the device, so any unbounded operation can stall it past the deadline.

## How to Apply

1. Render in fixed-size chunks (a "render quantum") sized to comfortably fit inside the buffer the device drains; everything the callback touches must already be allocated and resident.
2. Pre-touch and pin all buffers and voice state at startup so the first callback never page-faults; never grow a data structure from inside the callback.
3. If running a dedicated render thread rather than rendering inside the device callback, give it elevated priority and have it sleep on the device's "buffer low" event, waking only to top up the queue.

## Example

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

## Gotchas

- The first touch of a fresh page faults; lazily-allocated buffers underrun on their first use even though "nothing changed." Write to every buffer once at init.
- Sizing the queue-ahead target too tight invites starvation from ordinary OS scheduling jitter and high-latency devices (USB/Bluetooth audio); too loose adds latency. Tune it against the worst real device, not the dev machine.

## Related

[references/command-handoff.md](./command-handoff.md), [references/mixing-and-buffers.md](./mixing-and-buffers.md), **lock-free-guide**, **memory-management-guide**
