# mediabunny: Media Metadata and Frame Extraction

Use Mediabunny in `calculateMetadata` for `getVideoDuration()` (seconds → `* fps` frames), `getVideoDimensions()` (`{width, height}`), and `getAudioDuration()`. Verify codec support with `canDecode()` (Input + `getPrimaryVideoTrack`/audio track) before playback. Extract frames with `extractFrames()` via `VideoSampleSink`; pass `signal: controller.signal` to cancel.

```tsx
import {getVideoDimensions, getVideoDuration} from "mediabunny";

calculateMetadata: async ({props}) => {
  const duration = await getVideoDuration(props.videoUrl);
  const {width, height} = await getVideoDimensions(props.videoUrl);
  return {durationInFrames: Math.ceil(duration * 30), width, height};
};
```
