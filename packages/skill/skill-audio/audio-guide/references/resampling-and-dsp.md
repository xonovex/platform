# resampling-and-dsp: Sample-Rate Conversion, Pitch, and Per-Voice DSP

## Guideline

Resample every source to the mixer's internal rate by reading it with a fractional, per-voice playback step and interpolating between source samples; one knob — the step size — gives both rate conversion and pitch/speed control.

## Rationale

A source recorded at one rate cannot be summed against a mixer running at another rate without re-deriving its values at the mixer's sample instants. The clean abstraction is a sampler that, given a fractional time offset, returns the source's value there; advancing that offset by `src_rate / mix_rate` per output sample converts the rate, and scaling the step changes pitch and duration together (step 2.0 plays twice as fast, one octave up). Because real positions land between stored samples, you must interpolate; linear interpolation is cheap and usually adequate, with higher-order kernels (e.g. Hermite/cubic) available when fidelity matters. Doing this per voice lets each sound have its own pitch without touching the others.

## How to Apply

1. Define a sampler callback per source type: given a fractional offset and a count, it fills a mixer-rate buffer — `(user_data, offset, out, n, mix_rate)`. WAV, synth, and stream sources all hide behind this one signature.
2. Maintain a fractional read cursor per voice. Advance it by `step = (src_rate / mix_rate) * pitch` each output sample; `pitch` is the per-voice speed/pitch control.
3. Interpolate at the fractional position: linear blend of the two bracketing source samples for the default path; swap in a cubic/Hermite kernel only where the quality is worth the cost.
4. Keep DSP per-voice and bounded: pitch via step, gain/pan via the mix matrix; defer reverb and shared effect buses to a separate, fixed-cost post-mix stage.
5. Resample on load if the source is small and the rate fixed (one-time cost), or resample on the fly inside the sampler if the rate varies or the asset is streamed.

## Example

```c
// Sampler contract: produce `n` mixer-rate samples starting at `offset` seconds.
typedef uint32_t sampler_fn(const void *ud, double offset, float *out,
                            uint32_t n, float mix_rate);

// Linear-interpolating WAV sampler with per-voice pitch baked into the step.
uint32_t wav_sample(const void *ud, double offset, float *out,
                    uint32_t n, float mix_rate) {
  const wav_t *w = ud;
  double pos = offset * w->src_rate;            // source-sample position
  double step = (w->src_rate / mix_rate) * w->pitch; // pitch>1 => faster/higher
  for (uint32_t i = 0; i < n; ++i, pos += step) {
    uint32_t i0 = (uint32_t)pos;
    if (i0 + 1 >= w->num_samples) return i;     // ran out of source this block
    float frac = (float)(pos - i0);
    out[i] = w->samples[i0] * (1.0f - frac) + w->samples[i0 + 1] * frac; // lerp
  }
  return n;
}
```

## Gotchas

- Linear interpolation is a cheap default but rolls off highs and adds aliasing on big pitch-ups; reach for a cubic/Hermite or band-limited kernel only where the artifact is audible, not everywhere.
- Pitching a voice up makes it consume source samples faster, so it ends sooner — couple the read cursor to the loop/length logic or a sped-up loop will glitch at its seam.
- Accumulating the fractional position with `float` drifts over long sounds; keep the cursor in `double` (or fixed-point) so minutes-long sources stay in tune.
- Pitch-up reads past the source faster than pitch-1; always range-check the bracketing index, especially the `i0 + 1` neighbor at the very end of the buffer.
- Resampling on the audio thread is fine as long as it stays bounded and allocation-free; do not lazily allocate a scratch buffer there (see the callback contract).

## Related

[references/mixing-and-buffers.md](./mixing-and-buffers.md), [references/voice-management.md](./voice-management.md)
