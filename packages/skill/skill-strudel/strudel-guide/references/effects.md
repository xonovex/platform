# effects: Audio Effects and Processing

Chain effects with dot notation; mind gain staging.

```javascript
note("c2*8")
  .sound("sawtooth")
  .lpf(sine.range(300, 1500).slow(4)) // sweep filter
  .lpq(12) // high resonance (acid)
  .gain(0.5);
```

- Gain staging: kick 0.7-0.85, bass 0.5-0.6, leads 0.4-0.6, hats 0.2-0.35
- `.lpf(cutoff)` `.hpf(cutoff)`: shape frequencies (sub-bass 40-100 Hz, mud 300-600 Hz)
- `.lpq(resonance)`: peak at cutoff for acid sweeps
- `.room(0.5)` `.delay(0.3)`: space without muddying the mix
- `.shape(0.3)` `.crush(8)`: distortion and bit-reduction
- `.attack(0.1).sustain(0.5).release(0.3)`: envelope shape
