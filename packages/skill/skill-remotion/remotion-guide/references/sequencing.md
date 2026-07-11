# sequencing: Sequencing and Scene Orchestration

`<Sequence from durationInFrames>` shows content over a frame range and resets `useCurrentFrame()` to 0 inside it. Add `premountFor` to preload and avoid pop-in; `layout="none"` skips the `AbsoluteFill` wrapper (for flex/grid/canvas). Use `<Series>` for automatic sequential timing; a negative `offset` on `<Series.Sequence>` overlaps the previous scene.

```tsx
<Series>
  <Series.Sequence durationInFrames={2 * fps}>
    <Scene1 />
  </Series.Sequence>
  <Series.Sequence durationInFrames={3 * fps} offset={-0.5 * fps}>
    <Scene2 />
  </Series.Sequence>
</Series>
```
