---
name: strudel-guidelines
description: >-
  Trigger on Strudel music generation, `.strudel` files, live coding music tasks. Use when generating algorithmic music with Strudel.cc. Apply for pattern composition, effects processing, genre-specific production, browser-based live coding. Keywords: Strudel, TidalCycles, live coding, algorithmic music, mini-notation, patterns, electronic music, trance, ambient, techno.
---

# Strudel Music Coding Guidelines

## Requirements

- Browser with https://strudel.cc/ (no installation needed)
- Copy-paste code, Ctrl+Enter to play, Ctrl+. to stop

## Essentials

- **Tempo** - Set BPM first with `setcpm(BPM/4)`, see [reference/tempo-timing.md](reference/tempo-timing.md)
- **Mini-notation** - Space-separated sequences, `*` multiply, `/` divide, `<>` alternate, `[]` subdivide, `~` rest, see [reference/mini-notation.md](reference/mini-notation.md)
- **Layering** - Use `stack()` to combine patterns, each layer independent, see [reference/layering.md](reference/layering.md)
- **Sounds** - Drums: `bd`, `sd`, `hh`, `oh`, `cp`; Synths: `sine`, `sawtooth`, `square`, `triangle`, see [reference/sounds.md](reference/sounds.md)
- **Effects** - `.gain()`, `.lpf()`, `.room()`, `.delay()`, `.shape()`, `.crush()`, see [reference/effects.md](reference/effects.md)
- **Modulation** - `sine.range(a,b).slow(n)` for sweeps, `perlin` for organic variation, see [reference/modulation.md](reference/modulation.md)
- **Conditional** - `.every(n, fn)`, `.sometimes(fn)`, `.degradeBy(p)` for variation, see [reference/conditionals.md](reference/conditionals.md)
- **Structure** - Use `.mask()` for sections, `arrange()` for song form, see [reference/arrangement.md](reference/arrangement.md)

## Example

```javascript
setcpm(138 / 4);

stack(
  s("bd*4").gain(0.8),
  s("~ cp ~ cp").gain(0.5).room(0.3),
  s("hh*16").gain(0.3).lpf(sine.range(800, 4000).slow(4)),
  note("a2*8").sound("sawtooth").lpf(700).gain(0.5),
  note("<a3 f3 c4 g3>").sound("sawtooth").lpf(2000).room(0.8).gain(0.3),
);
```

## Progressive Disclosure

- Read [reference/mini-notation.md](reference/mini-notation.md) - When learning pattern syntax or fixing notation errors
- Read [reference/sounds.md](reference/sounds.md) - When selecting instruments or troubleshooting missing sounds
- Read [reference/effects.md](reference/effects.md) - When applying audio processing or creating textures
- Read [reference/modulation.md](reference/modulation.md) - When adding movement and dynamic variation
- Read [reference/conditionals.md](reference/conditionals.md) - When creating algorithmic variation
- Read [reference/layering.md](reference/layering.md) - When combining multiple patterns
- Read [reference/tempo-timing.md](reference/tempo-timing.md) - When setting tempo or working with time
- Read [reference/arrangement.md](reference/arrangement.md) - When structuring full compositions
- Read [reference/genre-trance.md](reference/genre-trance.md) - When producing uplifting/hard trance
- Read [reference/genre-ambient.md](reference/genre-ambient.md) - When creating atmospheric soundscapes
- Read [reference/genre-harsh.md](reference/genre-harsh.md) - When making experimental/industrial music
- Read [reference/scales-harmony.md](reference/scales-harmony.md) - When working with melodies and chords
