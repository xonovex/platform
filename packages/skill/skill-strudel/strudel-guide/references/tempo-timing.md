# tempo-timing: Tempo and Timing Control

Set tempo first with `setcpm(BPM/4)`, then shift timing with `.slow()` `.fast()` `.early()` `.late()`.

```javascript
setcpm(138 / 4); // 138 BPM
stack(
  s("bd*4"),
  s("sd*3").slow(4 / 3), // 3-per-cycle polyrhythm
  note("c4").slow(4).room(0.8), // 4-cycle pad
);
```

- `setcpm(BPM/4)`: 128=house, 138=trance, 170=drum&bass
- `.slow(n)`: pattern spans n cycles · `.fast(n)`: n repeats per cycle
- `.early(0.125)` `.late(0.125)`: shift by a fraction of a cycle
- `.slow(4/3)`: 3-per-cycle over a 4-beat backdrop
- `.swing(0.2)`: medium swing feel
