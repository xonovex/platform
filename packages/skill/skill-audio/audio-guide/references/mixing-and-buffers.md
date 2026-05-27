# mixing-and-buffers: Mixing Voices Into the Output Buffer

**Guideline:** Mix in a single internal format (deinterleaved 32-bit float, fixed internal sample rate), accumulate every active voice into the output channels through a per-voice gain matrix, ramp all gain changes over many samples, and scale to fit before converting to the device's integer format.

**Rationale:** A float accumulator with a 23-bit mantissa matches 24-bit integer precision while tolerating sums far outside [-1, 1] without the wraparound that destroys an integer mix — so you can add many voices and decide about loudness afterward. Routing each input channel to each output channel through a gain matrix `m[ic][oc]` generalizes volume, panning, and spatialization into one operation, which also vectorizes cleanly. The dangerous part is discontinuity: instantaneously changing a gain (start, stop, volume jump, pan snap) inserts a step into the waveform that the ear hears as a click or pop, so every coefficient change must be ramped across samples, not applied at once.

**How to Apply:**

1. Pick one internal format and convert everything into it on load/decode: deinterleaved (planar) float per channel, one fixed mix rate (e.g. 44.1 kHz). Keep the output buffer per-channel, not interleaved, while mixing.
2. For each active voice, for each of its input channels, for each output channel, do `out[oc][i] += voice_sample[ic][i] * m[ic][oc]` across the block; `m` is the per-voice gain/pan/3D matrix.
3. Vectorize the inner accumulate with a fused multiply-add over 8 lanes; mark the source/dest pointers `restrict` so the compiler can keep them in registers, and never alias the accumulator with a source.
4. Ramp every coefficient toward its target by a small fixed step per sample so a full 0->1 change spans tens of milliseconds (~70 ms is inaudible); apply the ramp inside the mix loop, not once per block.
5. After all voices are summed, if the peak exceeds the output range, scale the whole mix down to fit, then convert float -> the device's integer format and interleave for output.

**Example:**

```c
// Fused multiply-add accumulate of one source channel into one output channel.
// res and a/b must not alias; ping-pong the accumulator to satisfy `restrict`.
void mul_add(float *restrict res, const float *restrict acc,
             const float *restrict src, float gain, uint32_t n) {
  for (uint32_t i = 0; i < n; ++i)
    res[i] = acc[i] + src[i] * gain; // compiles to vfmadd over 8 lanes under AVX
}

// Click-free gain: step the live coefficient toward its target each sample.
// MAX_STEP chosen so a 0->1 change takes ~70 ms => 1/(0.070 * 44100).
#define GAIN_MAX_STEP (1.0f / 3072.0f)
static inline float ramp(float cur, float target) {
  float d = target - cur;
  if (d > GAIN_MAX_STEP) return cur + GAIN_MAX_STEP;
  if (d < -GAIN_MAX_STEP) return cur - GAIN_MAX_STEP;
  return target;
}
```

**Gotchas:**

- Snapping a gain to its new value between blocks is the most common click source; the fix is to ramp, and to keep ramping a "stopping" voice all the way to zero before you actually free it — never cut it dead.
- Interleaved I/O is the device's format, not the mixer's: mix planar, interleave only at the final conversion, or the per-channel inner loops stop vectorizing and the gain matrix gets awkward.
- Integer mixing overflows silently and wraps to a loud burst; float accumulation is what lets multiple loud voices coexist until you decide how to fit them.
- A `restrict` pointer that secretly aliases the accumulator produces wrong sums under optimization with no warning; ping-pong read/write buffers so source and destination are provably distinct.
- Scaling the whole mix to fit lets a loud sound duck everything else (acceptable), but it is not the same as proper limiting/compression; reach for a compressor only if the ducking is objectionable.
- Per-block ramping (one step per block instead of per sample) reintroduces stair-step clicks at block boundaries; the ramp must advance every sample.

**Related:** [references/resampling-and-dsp.md](./resampling-and-dsp.md), [references/voice-management.md](./voice-management.md), [references/audio-callback-thread.md](./audio-callback-thread.md), **data-oriented-design-guide**
