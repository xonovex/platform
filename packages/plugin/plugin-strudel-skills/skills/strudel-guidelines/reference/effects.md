# effects: Audio Effects and Processing

**Guideline:** Chain effects with dot notation; balance gain staging, use filters for EQ, and add space with reverb/delay.

**Rationale:** Tasteful effects shape tone and add movement; poor gain staging causes clipping or loss of clarity.

**Example:**

```javascript
// Acid bass
note("c2*8")
  .sound("sawtooth")
  .lpf(sine.range(300, 1500).slow(4)) // Sweep filter
  .lpq(12) // High resonance
  .gain(0.5);
```

**Techniques:**

- Gain staging: kick 0.7-0.85, bass 0.5-0.6, leads 0.4-0.6, hats 0.2-0.35
- `.lpf(cutoff)`, `.hpf(cutoff)`: Shape frequency space (sub-bass 40-100Hz, muddy 300-600Hz)
- `.lpq(resonance)`: Add peak at filter cutoff for acid sweeps
- `.room(0.5)`, `.delay(0.3)`: Create space without muddying mix
- `.shape(0.3)`, `.crush(8)`: Distortion and bit-reduction for texture
- `.attack(0.1).sustain(0.5).release(0.3)`: Control envelope shape
