# arrangement: Song Structure and Arrangement

Sequence sections with `arrange([cycles, pattern], ...)` and gate parts with `.mask()` to build tension, peak, and release.

```javascript
arrange([4, intro], [4, buildup], [8, drop], [4, breakdown], [8, drop]);
```

- `.mask("1 1 0 0 0 0 0 0")`: play only the first quarter of an 8-subdivision cycle
- `.every(n, x => x.fast(2))`: accelerate rhythms to build intensity
- `saw.range(200, 3000).slow(16)`: gradual filter build across a section
- Classic EDM form: intro → buildup → drop → breakdown → buildup → drop → outro
