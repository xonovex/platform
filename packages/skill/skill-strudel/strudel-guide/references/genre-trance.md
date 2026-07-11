# genre-trance: Uplifting/Hard Trance

Four-on-the-floor kicks, rolling bass, and filter sweeps at 138-145 BPM.

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

- Tempo `setcpm(138-145/4)` (classic to hard trance)
- Kick `s("bd*4").lpf(150)`; rolling bass `*8`/`*16` with `.lpf(sine.range(...))`
- Filter build: `.lpf(saw.range(200, 2000).slow(16))`
- Gain hierarchy: kick 0.85, bass 0.52, lead 0.42, pads 0.28, hats 0.28
