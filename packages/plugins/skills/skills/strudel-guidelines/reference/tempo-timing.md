# tempo-timing: Tempo and Timing Control

**Guideline:** Set tempo with `setcpm(BPM/4)` first, then use `.slow()`, `.fast()`, `.early()`, `.late()` for timing control.

**Rationale:** Tempo anchors rhythmic relationships; timing modifiers create polyrhythms, syncopation, and groove feels.

**Example:**
```javascript
setcpm(138/4)  // 138 BPM (uplifting trance)
stack(
  s("bd*4"),                    // Main beat
  s("sd*3").slow(4/3),          // 3-per-cycle polyrhythm
  note("c4").slow(4).room(0.8)  // 4-cycle pad
)
```

**Techniques:**
- `setcpm(BPM/4)`: Set tempo (128=house, 138=trance, 170=drum&bass)
- `.slow(n)`: Pattern takes n cycles (half speed)
- `.fast(n)`: Pattern plays n times per cycle (double speed)
- `.early(0.125)`, `.late(0.125)`: Shift timing by fractions of cycle
- Polyrhythms: `.slow(4/3)` creates 3-per-cycle over 4-beat backdrop
- `.swing(0.2)`: Add medium swing feel
