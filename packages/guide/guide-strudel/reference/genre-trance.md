# filename: genre-trance

**Guideline:** Build uplifting trance with four-on-the-floor kicks, rolling basslines, euphoric melodies, and filter sweeps at 138-145 BPM.

**Rationale:** Trance requires precise tempo, layered elements, and evolving filters; proper gain hierarchy and chord progressions maintain energy and emotional impact.

**Example:**

```javascript
setcpm(138 / 4);
stack(
  s("bd*4").gain(0.85).lpf(150),
  note("<a2 a2 f2 g2>*4")
    .sound("sawtooth")
    .lpf(sine.range(400, 1200).slow(4))
    .gain(0.52),
  s("hh*16").gain(0.28).hpf(6000),
);
```

**Techniques:**

- Tempo: `setcpm(138-145/4)` for classic to hard trance pacing
- Four-on-the-floor: `s("bd*4")` with punchy filtering `.lpf(150)`
- Rolling bass: `note("*8" or "*16")` with filter automation `.lpf(sine.range())`
- Filter sweeps: `.lpf(saw.range(200, 2000).slow(16))` for build tension
- Gain hierarchy: Kick 0.85, Bass 0.52, Lead 0.42, Pads 0.28, Hats 0.28
