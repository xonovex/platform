# charts: Animated Data Visualizations

Build charts as React/SVG and animate with `useCurrentFrame()` + `interpolate()`/`spring()`. Disable the library's own animation (`animation: false` in Chart.js; no D3 `.transition()`) — it bypasses frame control and desyncs.

- Line draw: SVG path `strokeDasharray`/`strokeDashoffset`
- Counter: `interpolate(frame, [0, fps], [0, target])`
- Stagger: offset per element by `index * delay` frames

```tsx
const progress = spring({
  frame: frame - i * 5 - 10,
  fps,
  config: {damping: 18},
});
const barHeight = (item.value / maxValue) * 300 * progress;
```
