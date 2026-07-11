# conditionals: Conditional and Probabilistic Transformations

Add algorithmic variation with periodic and probabilistic transforms instead of hand-editing patterns.

```javascript
s("bd*4")
  .every(4, (x) => x.fast(2)) // double speed every 4 cycles
  .every(8, (x) => x.gain(0.5))
  .sometimes((x) => x.speed(0.5)); // occasional pitch drop
```

- `.every(n, fn)`: apply `fn` every n cycles
- `.sometimes()` `.often()` `.rarely()` `.almostAlways()`: 50% / 75% / 10% / 90% probability
- `.degradeBy(0.7)`: drop 70% of events (sparse/glitchy)
- `.jux(fn)`: apply `fn` to the right channel only (stereo width)
- `.stut(4, 0.5, 1/16)`: echo/stutter with decay
- `.scramble()` `.shuffle()`: randomize event order / rotate subdivisions
