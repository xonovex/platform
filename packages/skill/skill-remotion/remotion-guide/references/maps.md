# maps: Map Animations with Mapbox

Disable Mapbox's own animation (`interactive: false`, `fadeDuration: 0`) and drive camera/lines from Remotion frames; wait for `map.load` and camera `idle` via `delayRender`/`continueRender`. Animate routes with turf.js (`turf.along`, `lineSliceAlong`). Render with `--gl=angle --concurrency=1` for Mapbox stability.

```tsx
const map = new Map({
  container: ref.current!,
  interactive: false,
  fadeDuration: 0,
  zoom: 11.53,
  center: [6.5615, 46.0598],
});
// route point: turf.along(lineString, distance * progress)
```

- Free camera: `getFreeCameraOptions()`, `lookAtPoint()`
- Hide basemap features: `setConfigProperty("basemap", name, false)`
