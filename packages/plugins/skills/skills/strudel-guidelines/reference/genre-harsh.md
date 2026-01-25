# filename: genre-harsh

**Guideline:** Create harsh, experimental textures using extreme distortion, bit crushing, noise, and chaotic polyrhythms for intentional sonic destruction.

**Rationale:** Harsh music demands intentional degradation and dissonance; extreme effect values and dark scales create cohesive experimental aesthetics.

**Example:**
```javascript
stack(
  s("bd(7,8)").gain(0.9).crush(5).shape(0.4),
  note("[-12 -11 -10]*32").scale("A2:locrian").sound("sawtooth")
    .lpf(perlin.range(400, 2000)).crush(5).shape(0.6).gain(0.5)
)
```

**Techniques:**
- Extreme effects: `.crush(2-4)`, `.coarse(32-64)`, `.shape(0.8-2)` for maximum harshness
- Noise layers: `s("white")` with `.hpf()`, `.bpf()`, `.degradeBy()` for texture
- Dissonant scales: Locrian, Phrygian, chromatic clusters for dark tonality
- Glitch methods: `.degradeBy(0.7)`, `.scramble()`, `.stut(8, 0.5)` for chaos
- Polyrhythms: Competing time signatures `.bd(7,8)`, `.sd(5,8)`, `.hh(11,16)` for instability
