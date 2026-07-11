# media: Audio and Video Manipulation

Use `<Audio>`/`<Video>` from `remotion` (not native elements) for frame-synced playback. Trim with `startFrom`/`endAt` in frames (`fps * seconds`); `playbackRate` scales speed; `volume` takes a number (0-1) or a `(frame) => number` callback for fades; `loop` repeats. Get duration via `getVideoDuration()`/`getAudioDuration()` in `calculateMetadata`.

```tsx
<Video src={staticFile("video.mp4")} startFrom={fps * 2} endAt={fps * 10} volume={0.8} />
<Audio src={staticFile("audio.mp3")} volume={(f) => interpolate(f, [0, fps], [0, 0.5])} />
```
