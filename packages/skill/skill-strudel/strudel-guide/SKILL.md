---
name: strudel-guide
description: "Use when writing or editing browser-based live-coding music in Strudel.cc. Triggers on `.strudel` files and prompts about mini-notation, pattern composition, FX chains, or generating algorithmic / genre-specific music (techno, ambient, trance, etc.), even when the user doesn't say 'Strudel'."
---

# Strudel Music Coding Guidelines

## Requirements

- Browser with https://strudel.cc/ (no installation needed)
- Copy-paste code, Ctrl+Enter to play, Ctrl+. to stop

## Essentials

- **Tempo** - Set BPM first with `setcpm(BPM/4)`, see [references/tempo-timing.md](references/tempo-timing.md)
- **Mini-notation** - Space-separated sequences, `*` multiply, `/` divide, `<>` alternate, `[]` subdivide, `~` rest, see [references/mini-notation.md](references/mini-notation.md)
- **Layering** - Use `stack()` to combine patterns, each layer independent, see [references/layering.md](references/layering.md)
- **Sounds** - Drums: `bd`, `sd`, `hh`, `oh`, `cp`; Synths: `sine`, `sawtooth`, `square`, `triangle`, see [references/sounds.md](references/sounds.md)
- **Effects** - `.gain()`, `.lpf()`, `.room()`, `.delay()`, `.shape()`, `.crush()`, see [references/effects.md](references/effects.md)
- **Modulation** - `sine.range(a,b).slow(n)` for sweeps, `perlin` for organic variation, see [references/modulation.md](references/modulation.md)
- **Conditional** - `.every(n, fn)`, `.sometimes(fn)`, `.degradeBy(p)` for variation, see [references/conditionals.md](references/conditionals.md)
- **Structure** - Use `.mask()` for sections, `arrange()` for song form, see [references/arrangement.md](references/arrangement.md)

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

## Gotchas

- Pattern transformations are right-associative — `s("bd").fast(2).rev()` is read right-to-left; the order changes the result
- Time signatures bind to the cycle, not bars — switching cycle length mid-piece shifts everything downstream
- Sample rate of loaded samples affects pitch — uploading 48 kHz samples to a 44.1 kHz session shifts notes by a semitone
- `stack` adds voices in parallel; `cat` concatenates in series — confusing them silently doubles or halves the pattern duration

## Progressive Disclosure

- Read [references/mini-notation.md](references/mini-notation.md) - Load when learning pattern syntax or fixing notation errors
- Read [references/sounds.md](references/sounds.md) - Load when selecting instruments or troubleshooting missing sounds
- Read [references/effects.md](references/effects.md) - Load when applying audio processing or creating textures
- Read [references/modulation.md](references/modulation.md) - Load when adding movement and dynamic variation
- Read [references/conditionals.md](references/conditionals.md) - Load when creating algorithmic variation
- Read [references/layering.md](references/layering.md) - Load when combining multiple patterns
- Read [references/tempo-timing.md](references/tempo-timing.md) - Load when setting tempo or working with time
- Read [references/arrangement.md](references/arrangement.md) - Load when structuring full compositions
- Read [references/genre-trance.md](references/genre-trance.md) - Load when producing uplifting/hard trance
- Read [references/genre-ambient.md](references/genre-ambient.md) - Load when creating atmospheric soundscapes
- Read [references/genre-harsh.md](references/genre-harsh.md) - Load when making experimental/industrial music
- Read [references/scales-harmony.md](references/scales-harmony.md) - Load when working with melodies and chords
