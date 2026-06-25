---
name: audio-guide
description: "Use when building a low-level, real-time audio/sound system that does its own mixing: the OS audio callback / render thread, summing voices into the output buffer, sample-rate conversion and per-voice pitch, voice pools and stealing, and handing play/stop/parameter changes from the game thread to the audio thread. Triggers on audio underrun/dropout, clicks/pops, the audio callback can't lock or allocate, buffer size vs latency, resampling/interpolation, mixing gain/pan/clipping, fixed voice pools, even when the user doesn't say 'audio'."
---

# Low-level Audio Guidelines (Real-Time Software Mixing)

Engine-agnostic architecture for a software audio mixer: a real-time render thread feeds the device, voices are summed into the output buffer, and the game thread talks to the audio thread only through a lock-free queue. The lock-free queue itself is owned by **lock-free-guide**; preallocated pools by **memory-management-guide**; struct-of-arrays voice state by **data-oriented-design-guide**.

## Essentials

- **The callback is hard real-time** - No locks, allocations, blocking, or syscalls on the audio path; bounded work per block, see [references/audio-callback-thread.md](references/audio-callback-thread.md)
- **Mix in float, ramp every gain** - Accumulate voices in deinterleaved float through a gain matrix; never snap a gain or you click, see [references/mixing-and-buffers.md](references/mixing-and-buffers.md)
- **One step does resample + pitch** - Read each source with a fractional, per-voice playback step and interpolate, see [references/resampling-and-dsp.md](references/resampling-and-dsp.md)
- **Fixed voice pool** - Preallocate `MAX_VOICES`; cap concurrency, steal by priority, ramp before reusing a slot, see [references/voice-management.md](references/voice-management.md)
- **Talk to the audio thread by message** - Post immutable commands through an SPSC ring; the audio thread owns voice state, see [references/command-handoff.md](references/command-handoff.md)

## The render path

- **Render quantum** - Produce fixed-size blocks sized to fit the device buffer; bounds per-block cost, see [references/audio-callback-thread.md](references/audio-callback-thread.md)
- **Queue ahead, sleep on the event** - Top up ~1.5 quanta, then wait on the device's buffer-low event, see [references/audio-callback-thread.md](references/audio-callback-thread.md)
- **Latency vs buffer size** - Bigger buffers resist underrun but add latency; tune against the worst real device, see [references/audio-callback-thread.md](references/audio-callback-thread.md)

## Mixing and signal

- **Gain matrix `m[ic][oc]`** - Volume, pan, and spatialization are all one per-voice input->output matrix, see [references/mixing-and-buffers.md](references/mixing-and-buffers.md)
- **Fit before output** - Scale the summed float mix to range, then convert/interleave to the device's integer format, see [references/mixing-and-buffers.md](references/mixing-and-buffers.md)
- **Interpolate fractional reads** - Linear by default, cubic/Hermite where fidelity is worth it, see [references/resampling-and-dsp.md](references/resampling-and-dsp.md)

## Voices and control

- **Steal the least important** - Full pool: cut the lowest-priority/quietest/oldest voice, never refuse silently, see [references/voice-management.md](references/voice-management.md)
- **Handles, not pointers** - Hand out slot+generation so stale references to recycled voices are detectable, see [references/voice-management.md](references/voice-management.md)
- **Commands are value-complete** - A command must not point at game-thread-mutable state, see [references/command-handoff.md](references/command-handoff.md)

## Gotchas

- An underrun is instantly, harshly audible: one lock the game thread holds, one `malloc` slow path, or one first-touch page fault inside the callback and you get a click or dropout — far worse than a dropped render frame.
- Snapping any gain (start, stop, volume, pan, steal) inserts a step into the waveform that clicks; ramp every coefficient over tens of milliseconds, and keep ramping a stopping voice to zero before freeing its slot.
- Integer mixing overflows and wraps to a loud burst; accumulate in float so many loud voices coexist, then scale to fit at the end.
- A "stopped" voice still occupies its slot until its fade completes, so the free count lags — size the pool for that overlap, not just steady state.
- Accumulating a voice's fractional read position in `float` drifts a long sound out of tune; keep the cursor in `double` or fixed-point.
- A full command queue must never block the audio thread and must never `malloc` a bigger one; pick a producer-side drop/coalesce policy up front.
- Sizing the queue-ahead too tight starves on ordinary scheduler jitter and high-latency USB/Bluetooth devices; too loose adds latency you can hear in interactive sounds.
- This handoff is one producer, one consumer; multiple game threads posting audio commands need an MPSC queue or a single funnel thread (see lock-free-guide).

## Progressive Disclosure

- Read [references/audio-callback-thread.md](references/audio-callback-thread.md) - Load when the audio callback/render thread underruns, glitches, or you need its no-lock/no-alloc contract, render quantum, or buffer-size/latency tuning
- Read [references/mixing-and-buffers.md](references/mixing-and-buffers.md) - Load when summing voices, handling channels/format, debugging clicks/pops, or fitting/clipping the mix
- Read [references/resampling-and-dsp.md](references/resampling-and-dsp.md) - Load when converting sample rates, interpolating, or implementing per-voice pitch/playback speed
- Read [references/voice-management.md](references/voice-management.md) - Load when allocating/stealing voices, sizing the voice pool, or managing voice lifetime and handles
- Read [references/command-handoff.md](references/command-handoff.md) - Load when passing play/stop/parameter changes from the game thread to the audio thread without locks
