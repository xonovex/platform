# transitions: Scene Transitions

**Guideline:** Use `<TransitionSeries>` from `@remotion/transitions` for fullscreen scene transitions.

**Rationale:** Transitions overlap scenes, reducing total composition duration. `<TransitionSeries>` handles complex timing automatically.

**Example:**
```tsx
import { TransitionSeries } from '@remotion/transitions'
import { linearTiming, springTiming } from '@remotion/transitions'
import { fade, slide } from '@remotion/transitions'
import { useVideoConfig } from 'remotion'

export const MyComposition = () => {
  const { fps } = useVideoConfig()

  return (
    <TransitionSeries>
      <TransitionSeries.Sequence durationInFrames={2 * fps}>
        <Scene1 />
      </TransitionSeries.Sequence>
      <TransitionSeries.Transition
        timing={linearTiming({ durationInFrames: 0.5 * fps })}
        effect={fade()}
      />
      <TransitionSeries.Sequence durationInFrames={3 * fps}>
        <Scene2 />
      </TransitionSeries.Sequence>
      <TransitionSeries.Transition
        timing={springTiming({ durationInFrames: 0.75 * fps })}
        effect={slide({ direction: 'from-left' })}
      />
      <TransitionSeries.Sequence durationInFrames={2 * fps}>
        <Scene3 />
      </TransitionSeries.Sequence>
    </TransitionSeries>
  )
  // Total: (2 + 0.5 + 3 + 0.75 + 2) * fps frames
}
```

**Techniques:**
- Install `@remotion/transitions` package before use
- Wrap all scenes with `<TransitionSeries>` container
- Use `<TransitionSeries.Sequence>` for each scene's content
- Add `<TransitionSeries.Transition>` between sequences for effects
- Account for overlap in duration math: `total = scene1 + scene2 - transition`
- Choose transition effect: `fade()`, `slide({direction})`, `wipe({direction})`, `flip({direction})`, `clockWipe()`
- Set timing with `linearTiming({durationInFrames})` for constant speed
- Use `springTiming({config, durationInFrames})` for animated spring transitions
- Calculate total duration accounting for overlapping transitions
- Test duration math carefully to ensure smooth transitions
