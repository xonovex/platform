# dom-measurement: Measuring DOM Elements

Divide `getBoundingClientRect()` values by `useCurrentScale()` — Remotion applies a scale transform, so raw measurements are scaled. Measure in a `useEffect` with `scale` as a dependency. Prefer `outline` over `border` for highlights: outline doesn't affect layout at different scales.

```tsx
const scale = useCurrentScale();
useEffect(() => {
  const rect = ref.current.getBoundingClientRect();
  setDimensions({width: rect.width / scale, height: rect.height / scale});
}, [scale]);
```
