# filename: scales-harmony

**Guideline:** Use `.scale()` to constrain notes to a key and chord notation for harmonically coherent compositions with intentional emotional expression.

**Rationale:** Keys unify bass, chords, and melody; scales prevent dissonance and enable cohesive harmonic progression across all layers.

**Example:**
```javascript
note("<a1 f1 c2 g1>*2").scale("A3:minor").sound("sawtooth")
  .lpf(400).gain(0.5),
note("<a3'min f3'maj c3'maj g3'maj>").sound("triangle")
  .room(0.7).gain(0.3)
```

**Techniques:**
- Scale constraint: `note("0 2 4 6").scale("C3:major")` to stay in key
- Chord notation: `note("c3'maj")`, `note("c3'min7")`, `note("c3'aug")` for harmony
- Progressions: i-VI-III-VII (trance), I-IV-V-I (house), vi-IV-I-V (pop)
- Scale moods: Major (happy), Minor (sad), Phrygian (dark), Locrian (metal)
- Transposition: `.add(12)` octave up, `.add(7)` fifth up for melodic variation
