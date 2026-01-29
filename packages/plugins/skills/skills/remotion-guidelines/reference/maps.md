# maps: Map Animations with Mapbox

**Guideline:** Disable Mapbox animations (fadeDuration: 0, interactive: false); control camera/lines via Remotion frames; use turf.js for route.

**Rationale:** Mapbox animations conflict with frame-by-frame rendering; Remotion frame control ensures deterministic animation.

**Example:**

```tsx
const _map = new Map({
  container: ref.current!,
  interactive: false, // CRITICAL
  fadeDuration: 0, // CRITICAL
  zoom: 11.53,
  center: [6.5615, 46.0598],
});
// Animate with: turf.along(lineString, distance * progress)
```

**Techniques:**

- Map config: interactive: false, fadeDuration: 0 disable Mapbox animations
- delayRender/continueRender: Wait for map.load and camera.idle events
- turf.js: Route distance, lineSliceAlong for curved animation
- Add GeoJSON sources for lines, markers, labels
- Free camera: Use getFreeCameraOptions(), lookAtPoint() for control
- Render: --gl=angle --concurrency=1 for Mapbox stability
- Performance: Hide unwanted features with setConfigProperty("basemap", name, false)
