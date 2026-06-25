---
name: remotion-guide
description: "Use when building or editing programmatic video in Remotion. Triggers on Remotion projects, `remotion.config.ts`, composition files, and prompts about useCurrentFrame, interpolate, spring, Sequence, transitions, captions, audio, or rendering MP4/GIF from React, even when the user doesn't say 'Remotion'."
---

# Remotion Coding Guidelines

## Requirements

- Remotion ≥ 4, React ≥ 18, TypeScript.

## Essentials

- **Frame-driven animations** - Use `useCurrentFrame()` for all motion; CSS transitions and Tailwind `animate-*` classes are forbidden, see [references/animations.md](references/animations.md)
- **Timing** - Write in seconds, multiply by `fps` from `useVideoConfig()`; use `interpolate()` or `spring()`, see [references/timing.md](references/timing.md)
- **Sequencing** - Use `<Sequence>` with `from`/`durationInFrames`; always add `premountFor` to preload, see [references/sequencing.md](references/sequencing.md)
- **Assets** - Use `staticFile()` for public assets; use `<Img>`, `<Video>`, `<Audio>` components, see [references/assets.md](references/assets.md)
- **Compositions** - Define in `Root.tsx` with `type` (not interface) for props; use `calculateMetadata()` for dynamic values, see [references/compositions.md](references/compositions.md)
- **Transitions** - Use `<TransitionSeries>` for scene changes; duration overlaps reduce total length, see [references/transitions.md](references/transitions.md)
- **Text** - Load fonts via `@remotion/google-fonts`; measure with `@remotion/layout-utils`, see [references/text.md](references/text.md)

## Example

```tsx
import {
  AbsoluteFill,
  interpolate,
  spring,
  useCurrentFrame,
  useVideoConfig,
} from "remotion";

export function FadeIn() {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const opacity = interpolate(frame, [0, 0.5 * fps], [0, 1], {
    extrapolateRight: "clamp",
  });

  const scale = spring({
    frame,
    fps,
    config: {damping: 200},
  });

  return (
    <AbsoluteFill style={{opacity, transform: `scale(${scale})`}}>
      <h1>Hello Remotion</h1>
    </AbsoluteFill>
  );
}
```

## Gotchas

- Rendering is deterministic — any `Math.random()` or `Date.now()` produces different frames across renders; seed via `useCurrentFrame()` instead
- `durationInFrames` is configured per-composition; mismatching it with the timeline causes content cut-off or padding
- Audio sync requires `<Audio />` inside the composition timeline — playing audio in raw HTML elements desyncs from frame ticks
- Heavy components inside `<Series />` re-mount per segment — extract stable computation outside or memoize per-frame

## Progressive Disclosure

- Read [references/animations.md](references/animations.md) - Load when creating motion with interpolate, spring, or easing
- Read [references/timing.md](references/timing.md) - Load when converting seconds to frames or working with durations
- Read [references/sequencing.md](references/sequencing.md) - Load when orchestrating multiple elements or scenes
- Read [references/assets.md](references/assets.md) - Load when loading images, videos, audio, or fonts
- Read [references/compositions.md](references/compositions.md) - Load when defining video structure or dynamic metadata
- Read [references/transitions.md](references/transitions.md) - Load when adding scene transitions like fade, slide, wipe
- Read [references/text.md](references/text.md) - Load when animating text, measuring dimensions, or loading fonts
- Read [references/media.md](references/media.md) - Load when trimming, adjusting volume, or manipulating audio/video
- Read [references/captions.md](references/captions.md) - Load when adding subtitles or TikTok-style captions
- Read [references/three-d.md](references/three-d.md) - Load when integrating Three.js 3D content
- Read [references/charts.md](references/charts.md) - Load when building animated data visualizations
- Read [references/gifs.md](references/gifs.md) - Load when displaying GIFs, APNG, WebP, or AVIF animations
- Read [references/lottie.md](references/lottie.md) - Load when embedding Lottie animations
- Read [references/maps.md](references/maps.md) - Load when creating map animations with Mapbox
- Read [references/mediabunny.md](references/mediabunny.md) - Load when getting video duration, dimensions, or extracting frames
- Read [references/dom-measurement.md](references/dom-measurement.md) - Load when measuring DOM elements with getBoundingClientRect
