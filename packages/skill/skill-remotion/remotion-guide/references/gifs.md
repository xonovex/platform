# gifs: Animated Images (GIF, APNG, WebP, AVIF)

Use `<AnimatedImage>` from `remotion` (not native `<img>`) for frame-synced animated images; set explicit `width`/`height`. `loop` values: `loop` (default), `pause-after-finish`, `clear-after-finish`. `fit`: `fill` | `contain` | `cover`. `playbackRate` scales speed. Get duration for dynamic length with `getGifDurationInSeconds()`.

```tsx
<AnimatedImage
  src={staticFile("animation.gif")}
  width={1920}
  height={1080}
  playbackRate={1}
  loop
  fit="contain"
/>
```
