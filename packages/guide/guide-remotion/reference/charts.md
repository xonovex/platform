# filename: charts

**Guideline:** Build charts with React/SVG; animate with `useCurrentFrame()` + `interpolate()` or `spring()`; disable third-party library animations.

**Rationale:** Remotion controls frame-by-frame timing; library animations bypass frame control; manual rendering ensures sync with audio/video.

**Example:**

```tsx
function BarChart({data}) {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  return (
    <AbsoluteFill style={{backgroundColor: "#1a1a2e"}}>
      <svg width="100%" height="100%">
        {data.map((item, i) => {
          const delay = i * 5;
          const progress = spring({
            frame: frame - delay - 10,
            fps,
            config: {damping: 18},
          });
          const barHeight = (item.value / maxValue) * 300 * progress;
          return (
            <rect
              key={i}
              y={350 - barHeight}
              height={barHeight}
              fill={item.color}
            />
          );
        })}
      </svg>
    </AbsoluteFill>
  );
}
```

**Techniques:**

- Bar chart: `spring()` for per-bar stagger; `frame - delay` for sequential animation
- Pie chart: `interpolate()` for smooth arc fill; use SVG path math for arcs
- Line chart: SVG path with `strokeDasharray`/`strokeDashoffset` for drawing animation
- Counter numbers: `interpolate(frame, [0, fps], [0, target])` for smooth count
- Stagger elements: Use `index * delay` to offset animation start per element
- Disable library animations: Set `animation: false` in Chart.js; don't use D3 `.transition()`
- Use `interpolate()` for smooth ranges; `spring()` for bouncy/organic feel
