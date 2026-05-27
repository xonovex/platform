# Sources

## Game-engine development blog (archive)

- **URL:** https://archive-host.github.io/blog_archive/post/writing-a-low-level-sound-system/index.html
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Essentials, The render path, Mixing and signal, Voices and control, Gotchas
  - The whole architecture of a software audio mixer: real-time render thread, float mixing, resampling/pitch, voice handling, and game→audio thread handoff
- **Aspects extracted:**
  - "Writing a low-level sound system" — dedicated high-priority render thread, render quantum (257\*8 samples ~46 ms), queue-ahead vs starvation, waveOut/WASAPI buffer feeding, main-thread-vs-dedicated-thread latency, no-lock/no-alloc real-time discipline, page-fault avoidance, big-VM-allocation pointer bump → `references/audio-callback-thread.md`
  - "Writing a low-level sound system" — internal deinterleaved 32-bit float format at 44.1 kHz, per-voice gain matrix `m[ic][oc]`, AVX fused-multiply-add `mul_add` with `restrict` + ping-pong accumulator, gain ramping (1/3072 per sample ~70 ms) to kill clicks/pops, ramping a stopping voice to zero, scaling the mix to fit the output range / clipping → `references/mixing-and-buffers.md`
  - "Writing a low-level sound system" — sampler callback contract `(ud, offset, buffer, num_samples, sample_frequency)`, linear interpolation between WAV samples, pitch via sample-frequency / step scaling (octave-up at 2x), Hermite as a future higher-quality option → `references/resampling-and-dsp.md`
  - "Writing a low-level sound system" — software-mixing many sources, bounded per-block cost from a fixed source set, per-voice state, graceful stop via gain ramp before retiring a source → `references/voice-management.md`
  - "Writing a low-level sound system" — double-buffering all shared data with a single lock to swap so the main and render threads never touch the same data simultaneously (generalized here to a lock-free SPSC command handoff) → `references/command-handoff.md`

## Refresh Workflow

1. Re-read the upstream source(s) above
2. Diff against the prior pull (or scan for newly added sections)
3. For each changed area, update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
