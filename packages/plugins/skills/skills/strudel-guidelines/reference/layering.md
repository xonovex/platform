# filename: layering

**Guideline:** Use `stack()` to combine independent patterns organized by role (rhythm, bass, harmony, melody) with balanced gain levels.

**Rationale:** Clear layer hierarchy prevents overlap and mud; role-based organization enables independent editing and coherent mixes.

**Example:**
```javascript
stack(
  s("bd*4").gain(0.8).lpf(150),
  note("c2*8").sound("sawtooth").lpf(600).gain(0.5),
  note("c5 e5 g5 e5").sound("sawtooth").lpf(3000).gain(0.4),
  s("hh*16").gain(0.25).hpf(6000)
)
```

**Techniques:**
- Layer hierarchy: Kick 0.8-0.9, Bass 0.5-0.6, Lead 0.4-0.5, Pads 0.2-0.4, Hats 0.2-0.4
- Multi-layer drums: `.late(0.01)` and `.late(0.02)` offsets for depth
- Role organization: Rhythm → Bass → Harmony → Melody → Texture in stack order
- Variable patterns: Define as `const layer = pattern` for clarity and reusability
- Combination functions: `stack()` simultaneous, `cat()` sequential per cycle, `arrange()` explicit counts
