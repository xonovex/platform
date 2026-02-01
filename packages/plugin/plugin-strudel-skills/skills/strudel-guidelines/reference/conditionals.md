# conditionals: Conditional and Probabilistic Transformations

**Guideline:** Use `.every(n, fn)`, probability functions, and `.degradeBy()` to add algorithmic variation without manual editing.

**Rationale:** Procedural variation creates organic, evolving patterns; manual static edits sound lifeless and repetitive.

**Example:**

```javascript
// Evolving drum pattern
s("bd*4")
  .every(4, (x) => x.fast(2)) // Double speed every 4 cycles
  .every(8, (x) => x.gain(0.5)) // Quieter every 8 cycles
  .sometimes((x) => x.speed(0.5)); // Occasional pitch drop
```

**Techniques:**

- `.every(n, fn)`: Apply transformation periodically (repeats every n cycles)
- `.sometimes()`, `.often()`, `.rarely()`, `.almostAlways()`: Probabilistic 50%, 75%, 10%, 90%
- `.degradeBy(0.7)`: Keep only 70% of events (sparse/glitchy feel)
- `.jux(fn)`: Apply to right channel only (stereo width effect)
- `.stut(4, 0.5, 1/16)`: Echo/stutter with decay
- `.scramble()`, `.shuffle()`: Randomize event order or rotate subdivisions
