# modulation: Dynamic Parameter Modulation

Drive parameters with signal generators instead of static values for organic movement.

```javascript
note("c2*8")
  .sound("sawtooth")
  .lpf(sine.range(200, 2000).slow(4))
  .lpq(8)
  .gain(0.5);
```

- `sine.range(min, max).slow(n)`: smooth LFO oscillation
- `saw.range(min, max).slow(n)`: linear ramp for builds
- `perlin.range(min, max)`: smooth organic randomness (non-repetitive)
- `rand.range(min, max)`: true random per event (glitch)
- Common targets: `.lpf()` `.gain()` `.pan()` `.room()` `.lpq()`
- Stepped changes: `.crush("<3 4 5 6>")`, `.gain("<0.8 0.9 1 0.85>")`
