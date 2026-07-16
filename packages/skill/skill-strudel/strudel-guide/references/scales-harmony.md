# scales-harmony: Scales and Chords

Constrain notes with `.scale()` and use chord notation.

```javascript
stack(
  note("<a1 f1 c2 g1>*2")
    .scale("A3:minor")
    .sound("sawtooth")
    .lpf(400)
    .gain(0.5),
  note("<a3'min f3'maj c3'maj g3'maj>").sound("triangle").room(0.7).gain(0.3),
);
```

- Scale constraint: `note("0 2 4 6").scale("C3:major")` (numbers = scale degrees)
- Chord notation: `note("c3'maj")` `note("c3'min7")` `note("c3'aug")`
- Progressions: i-VI-III-VII (trance), I-IV-V-I (house), vi-IV-I-V (pop)
- Scale moods: major (bright), minor (sad), Phrygian (dark), Locrian (metal)
- Transpose: `.add(12)` octave up, `.add(7)` fifth up
