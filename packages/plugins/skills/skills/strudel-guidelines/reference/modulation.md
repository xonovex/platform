# modulation: Dynamic Parameter Modulation

**Guideline:** Replace static values with signal generators (sine, perlin, rand) to create organic evolution and movement.

**Rationale:** Dynamic modulation adds life and interest; static values sound robotic and unmotivated.

**Example:**
```javascript
// Classic filter sweep
note("c2*8").sound("sawtooth")
  .lpf(sine.range(200, 2000).slow(4))
  .lpq(8)
  .gain(0.5)
```

**Techniques:**
- `sine.range(min, max).slow(n)`: Smooth, predictable LFO oscillation
- `saw.range(min, max).slow(n)`: Linear ramps for gradual builds
- `perlin.range(min, max)`: Smooth organic randomness (non-repetitive feel)
- `rand.range(min, max)`: True random per event (glitch effects)
- Common targets: `.lpf()`, `.gain()`, `.pan()`, `.room()`, `.lpq()` (resonance)
- Cycle-based: `.crush("<3 4 5 6>")` or `.gain("<0.8 0.9 1 0.85>")` for stepped changes
