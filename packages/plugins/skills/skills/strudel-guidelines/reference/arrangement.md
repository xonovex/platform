# arrangement: Song Structure and Arrangement

**Guideline:** Organize compositions into distinct sections using `.mask()` and `arrange()` to build tension, peaks, and release.

**Rationale:** Structured arrangements create emotional arcs; static loops lack engagement and dynamics.

**Example:**
```javascript
arrange(
  [4, intro],     // 4 cycles
  [4, buildup],   // Building tension
  [8, drop],      // Climax
  [4, breakdown], // Release
  [8, drop]       // Return
)
```

**Techniques:**
- `.mask("1 1 0 0 0 0 0 0")`: Play first half of 8-subdivision cycle
- `arrange([cycles, pattern], ...)`: Sequence sections with precise lengths
- `.every(n, x => x.fast(2))`: Accelerate rhythms to build intensity
- Layer stacking: Add/remove instruments progressively for tension
- Filter sweeps: Use `saw.range(200, 3000).slow(16)` for gradual builds
- Classic EDM: intro→buildup→drop→breakdown→buildup→drop→outro
