# genre-ambient: Ambient and Atmospheric Music

Slow tempo, sparse elements, heavy reverb, and long envelopes for meditative soundscapes.

```javascript
setcpm(70 / 4); // 70 BPM
stack(
  note("c2")
    .sound("sine")
    .attack(3)
    .sustain(8)
    .release(4)
    .room(0.9)
    .gain(0.25)
    .slow(8),
  note("<c4 eb4 g4 bb4>")
    .sound("triangle")
    .attack(1)
    .sustain(2)
    .release(2)
    .room(0.85)
    .lpf(perlin.range(500, 2000).slow(8))
    .gain(0.2)
    .slow(2),
  note("g5 ~ ~ ~ ~ eb5 ~ ~")
    .sound("sine")
    .room(0.9)
    .delay(0.6)
    .delayfeedback(0.7)
    .gain(0.2)
    .degradeBy(0.4),
  s("white").lpf(perlin.range(300, 800).slow(16)).gain(0.08),
);
```

- Tempo 60-90 BPM; reverb `.room(0.85-0.95)`
- Long envelopes: `.attack(2-4)` `.release(3-4)`
- `.degradeBy(0.3-0.5)` to thin notes; stack sine/triangle drones across octaves with `.slow()`
