# layering: Combining Patterns with stack()

Combine independent patterns with `stack()`, ordered by role with balanced gains.

```javascript
stack(
  s("bd*4").gain(0.8).lpf(150),
  note("c2*8").sound("sawtooth").lpf(600).gain(0.5),
  note("c5 e5 g5 e5").sound("sawtooth").lpf(3000).gain(0.4),
  s("hh*16").gain(0.25).hpf(6000),
);
```

- Gain hierarchy: kick 0.8-0.9, bass 0.5-0.6, lead 0.4-0.5, pads 0.2-0.4, hats 0.2-0.4
- Stack order: rhythm → bass → harmony → melody → texture
- `.late(0.01)` / `.late(0.02)`: micro-offset layered drums for depth
- `stack()` simultaneous · `cat()` sequential per cycle · `arrange()` explicit cycle counts
