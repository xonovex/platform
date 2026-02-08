---
name: remotion-guidelines
description: Trigger on Remotion video projects, `remotion.config.ts`, video composition files. Use when creating programmatic videos with React. Apply for animations, sequencing, media handling, transitions, captions. Keywords: Remotion, video, React, animation, composition, useCurrentFrame, interpolate, spring, Sequence.
---

# Remotion Coding Guidelines

## Requirements

- Remotion ≥ 4, React ≥ 18, TypeScript.

## Essentials

- **Frame-driven animations** - Use `useCurrentFrame()` for all motion; CSS transitions and Tailwind `animate-*` classes are forbidden, see [reference/animations.md](reference/animations.md)
- **Timing** - Write in seconds, multiply by `fps` from `useVideoConfig()`; use `interpolate()` or `spring()`, see [reference/timing.md](reference/timing.md)
- **Sequencing** - Use `<Sequence>` with `from`/`durationInFrames`; always add `premountFor` to preload, see [reference/sequencing.md](reference/sequencing.md)
- **Assets** - Use `staticFile()` for public assets; use `<Img>`, `<Video>`, `<Audio>` components, see [reference/assets.md](reference/assets.md)
- **Compositions** - Define in `Root.tsx` with `type` (not interface) for props; use `calculateMetadata()` for dynamic values, see [reference/compositions.md](reference/compositions.md)
- **Transitions** - Use `<TransitionSeries>` for scene changes; duration overlaps reduce total length, see [reference/transitions.md](reference/transitions.md)
- **Text** - Load fonts via `@remotion/google-fonts`; measure with `@remotion/layout-utils`, see [reference/text.md](reference/text.md)

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

## Progressive Disclosure

- Read [reference/animations.md](reference/animations.md) - When creating motion with interpolate, spring, or easing
- Read [reference/timing.md](reference/timing.md) - When converting seconds to frames or working with durations
- Read [reference/sequencing.md](reference/sequencing.md) - When orchestrating multiple elements or scenes
- Read [reference/assets.md](reference/assets.md) - When loading images, videos, audio, or fonts
- Read [reference/compositions.md](reference/compositions.md) - When defining video structure or dynamic metadata
- Read [reference/transitions.md](reference/transitions.md) - When adding scene transitions like fade, slide, wipe
- Read [reference/text.md](reference/text.md) - When animating text, measuring dimensions, or loading fonts
- Read [reference/media.md](reference/media.md) - When trimming, adjusting volume, or manipulating audio/video
- Read [reference/captions.md](reference/captions.md) - When adding subtitles or TikTok-style captions
- Read [reference/three-d.md](reference/three-d.md) - When integrating Three.js 3D content
- Read [reference/charts.md](reference/charts.md) - When building animated data visualizations
- Read [reference/gifs.md](reference/gifs.md) - When displaying GIFs, APNG, WebP, or AVIF animations
- Read [reference/lottie.md](reference/lottie.md) - When embedding Lottie animations
- Read [reference/maps.md](reference/maps.md) - When creating map animations with Mapbox
- Read [reference/mediabunny.md](reference/mediabunny.md) - When getting video duration, dimensions, or extracting frames
- Read [reference/dom-measurement.md](reference/dom-measurement.md) - When measuring DOM elements with getBoundingClientRect
