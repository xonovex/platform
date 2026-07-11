# captions: Subtitles and TikTok-Style Captions

Parse SRT with `parseSrt({input})` from `@remotion/captions` (captions have `startMs`, `endMs`, `text`) rather than by hand. Use `createTikTokStyleCaptions({captions})` for word-by-word pages of tokens; check per-token active state to color/scale the current word. Match the active caption by `currentTimeMs = (frame / fps) * 1000`; fetch remote SRT inside `delayRender()`/`continueRender()`.

```tsx
const currentTimeMs = (frame / fps) * 1000;
const active = captions.find(
  (c) => currentTimeMs >= c.startMs && currentTimeMs < c.endMs,
);
```
