# assets: Static Files and Asset Loading

Reference `public/` assets with `staticFile()` (auto-encodes special chars) and use Remotion's `<Img>`, `<Video>`, `<Audio>` — they await full load, native `<img>`/`<video>`/`<audio>` don't. Remote URLs (CORS-enabled) work without `staticFile()`.

```tsx
<Img src={staticFile("logo.png")} />
<Video src={staticFile("video.mp4")} />
<Audio src={staticFile("music.mp3")} volume={0.5} />
const {fontFamily} = loadFont({family: "Custom", url: staticFile("fonts/font.woff2")});
```
