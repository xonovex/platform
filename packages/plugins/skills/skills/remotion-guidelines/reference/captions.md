# filename: captions

**Guideline:** Use `@remotion/captions` with `parseSrt()` for accurate timing; use `createTikTokStyleCaptions()` for word-by-word display.

**Rationale:** SRT parsing is error-prone manually; library handles edge cases; `TikTokStyleCaptions` enables smooth word highlighting matching audio timing.

**Example:**
```tsx
function Captions({captions}) {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  const currentTimeMs = (frame / fps) * 1000;

  const activeCaption = captions.find(
    (c) => currentTimeMs >= c.startMs && currentTimeMs < c.endMs
  );

  return activeCaption ? (
    <div style={{position: "absolute", bottom: 100, fontSize: 48, color: "white"}}>
      {activeCaption.text}
    </div>
  ) : null;
}
```

**Techniques:**
- Parse SRT: `const {captions} = parseSrt({input: srtContent})`; captions have `startMs`, `endMs`, `text`
- Frame to time: `currentTimeMs = (frame / fps) * 1000`; match against caption range
- TikTok style: `createTikTokStyleCaptions({captions})` returns pages with tokens (individual words)
- Word highlighting: Check `isActive` per token; apply color/scale transform on active words
- Remote SRT: Use `delayRender()`/`continueRender()` for async fetch; load before rendering
- Styling: Add background, padding, text-shadow for readability; center with `transform: translateX(-50%)`
- Dynamic timing: Use `spring()` for scale animation on word activation for emphasis effect
