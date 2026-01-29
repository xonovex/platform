# media: Audio and Video Manipulation

**Guideline:** Use `<Audio>` and `<Video>` from Remotion for frame-synchronized playback.

**Rationale:** Native HTML elements don't sync properly with Remotion's timeline. Remotion components provide frame-synchronized playback control.

**Example:**

```tsx
import {Audio, useVideoConfig, Video} from "remotion";

export const MediaComposition = () => {
  const {fps} = useVideoConfig();

  return (
    <>
      <Video
        src="video.mp4"
        startFrom={fps * 2} // Start at 2 seconds
        endAt={fps * 10} // End at 10 seconds
        playbackRate={1}
        volume={0.8}
        style={{width: "100%", height: "100%"}}
      />
      <Audio src="audio.mp3" volume={0.5} />
    </>
  );
};
```

**Techniques:**

- Import `Audio` and `Video` from `remotion` for frame-synced playback
- Use `staticFile()` for local media files, remote URLs for external sources
- Set static volume with number (0-1 range) or dynamic with callback function
- Trim playback: `startFrom` and `endAt` props in frames (use `fps` for seconds)
- Control speed with `playbackRate` prop (0.5 = half speed, 2 = double speed)
- Enable looping with `loop` prop for indefinite repetition
- Create fade effects with `interpolate()` and frame-based callbacks
- Get media duration with `getVideoDuration()` or `getAudioDuration()` in calculateMetadata
- Apply CSS styles: width, height, position, opacity, filters
