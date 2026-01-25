# mediabunny: Media Metadata and Frame Extraction

**Guideline:** Use Mediabunny for video/audio duration, dimensions, and frame extraction; check decode compatibility before playback.

**Rationale:** Duration enables accurate durationInFrames; decode checks prevent runtime failures with unsupported codecs.

**Example:**
```tsx
import {getVideoDuration, getVideoDimensions} from "mediabunny";

calculateMetadata: async ({props}) => {
  const duration = await getVideoDuration(props.videoUrl);
  const {width, height} = await getVideoDimensions(props.videoUrl);
  return {
    durationInFrames: Math.ceil(duration * 30),
    width, height
  };
}
```

**Techniques:**
- getVideoDuration(): Returns seconds; multiply by fps for frames
- getAudioDuration(): For audio files
- getVideoDimensions(): Returns {width, height} for dynamic sizing
- canDecode(): Check Input + getPrimaryVideoTrack/Audio before playback
- extractFrames(): Sample frames at specific timestamps using VideoSampleSink
- Filmstrip: Calculate timestamps based on duration and desired grid spacing
- Cancellation: Pass signal: controller.signal for AbortSignal support
