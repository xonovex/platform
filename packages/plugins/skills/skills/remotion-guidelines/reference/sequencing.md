# sequencing: Sequencing and Scene Orchestration

**Guideline:** Use `<Sequence>` for timed playback with local frame context and `premountFor` to prevent pop-in.

**Rationale:** Sequences provide local frame context reset and timing control. `premountFor` preloads components early to avoid rendering artifacts.

**Example:**

```tsx
import {Sequence, Series, useCurrentFrame, useVideoConfig} from "remotion";

export const MyComposition = () => {
  const {fps} = useVideoConfig();

  return (
    <Series>
      <Series.Sequence durationInFrames={2 * fps}>
        <Scene1 />
      </Series.Sequence>
      <Series.Sequence durationInFrames={3 * fps} offset={-0.5 * fps}>
        <Scene2 /> {/* Overlaps with Scene1 by 0.5 sec */}
      </Series.Sequence>
      <Series.Sequence durationInFrames={2 * fps}>
        <Scene3 />
      </Series.Sequence>
    </Series>
  );
};

const Scene1 = () => {
  const frame = useCurrentFrame(); // Starts at 0 in this Sequence
  return <div style={{opacity: frame / 30}} />;
};
```

**Techniques:**

- Use `<Sequence>` to show/hide components at specific frame ranges
- Set `from` prop for start frame (calculate with `fps`: `1.5 * fps` for 1.5 seconds)
- Set `durationInFrames` for length in frames (calculate with `fps`: `2 * fps` for 2 seconds)
- Enable `premountFor` to preload components early and avoid rendering pop-in
- Use `layout="none"` to skip AbsoluteFill wrapper for flex/grid/canvas layouts
- Remember `useCurrentFrame()` returns local frame context within Sequence (starts at 0)
- Use `<Series>` for automatic sequential timing of multiple scenes
- Create overlaps with `offset` prop on Series.Sequence (negative value for overlap)
- Nest Sequences for complex timing and layered compositions
