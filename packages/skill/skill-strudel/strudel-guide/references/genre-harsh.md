# genre-harsh: Experimental/Industrial Textures

Extreme distortion, bit-crushing, noise, and chaotic polyrhythms for intentional sonic destruction.

```javascript
stack(
  s("bd(7,8)").gain(0.9).crush(5).shape(0.4),
  note("[-12 -11 -10]*32")
    .scale("A2:locrian")
    .sound("sawtooth")
    .lpf(perlin.range(400, 2000))
    .crush(5)
    .shape(0.6)
    .gain(0.5),
);
```

- Extreme effects: `.crush(2-4)` `.coarse(32-64)` `.shape(0.8-2)`
- Noise: `s("white")` with `.hpf()` `.bpf()` `.degradeBy()`
- Dark scales: Locrian, Phrygian, chromatic clusters
- Glitch: `.degradeBy(0.7)` `.scramble()` `.stut(8, 0.5)`
- Competing polyrhythms: `bd(7,8)`, `sd(5,8)`, `hh(11,16)`
