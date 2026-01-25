# gifs: Animated Images (GIF, APNG, WebP, AVIF)

**Guideline:** Use `<AnimatedImage>` for frame-synchronized playback of animated images.

**Rationale:** Native img tags don't sync with Remotion's timeline. `<AnimatedImage>` provides proper frame synchronization and playback control.

**Example:**
```tsx
import { AnimatedImage, staticFile, Composition } from 'remotion'

export const MyComposition = () => (
  <Composition>
    <AnimatedImage
      src={staticFile('animation.gif')}
      width={1920}
      height={1080}
      playbackRate={1}
      loop
      fit="contain"
    />
  </Composition>
)
```

**Techniques:**
- Import `AnimatedImage` and `staticFile` from `remotion` package
- Set explicit `width` and `height` props for proper sizing and rendering
- Use `staticFile()` for local assets, remote URLs with CORS headers
- Control playback speed with `playbackRate` prop (0.5 = half, 2 = double)
- Set loop behavior: `loop` (default), `pause-after-finish`, or `clear-after-finish`
- Choose sizing mode: `fit="fill"` (stretch), `fit="contain"` (aspect-ratio), `fit="cover"` (crop)
- Get GIF duration with `getGifDurationInSeconds()` for dynamic composition length
- Apply CSS styles: borderRadius, position, opacity, filters
