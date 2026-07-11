# transitions: Scene Transitions

Wrap scenes in `<TransitionSeries>` from `@remotion/transitions`, alternating `<TransitionSeries.Sequence>` with `<TransitionSeries.Transition>`. Transitions overlap adjacent scenes, so total duration = sum of sequences − sum of transition durations.

- Effects: `fade()`, `slide({direction})`, `wipe({direction})`, `flip({direction})`, `clockWipe()`
- Timing: `linearTiming({durationInFrames})` (constant) or `springTiming({config, durationInFrames})`

```tsx
<TransitionSeries>
  <TransitionSeries.Sequence durationInFrames={2 * fps}>
    <Scene1 />
  </TransitionSeries.Sequence>
  <TransitionSeries.Transition
    timing={linearTiming({durationInFrames: 0.5 * fps})}
    effect={fade()}
  />
  <TransitionSeries.Sequence durationInFrames={3 * fps}>
    <Scene2 />
  </TransitionSeries.Sequence>
</TransitionSeries>
```
