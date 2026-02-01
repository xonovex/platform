# assets: Static Files and Asset Loading

**Guideline:** Use `staticFile()` for local assets in `public/` folder; use Remotion media components `<Img>`, `<Video>`, `<Audio>` that await loading.

**Rationale:** staticFile() ensures correct paths; Remotion components wait for full asset load before rendering (native HTML doesn't).

**Example:**

```tsx
<Img src={staticFile("logo.png")} />
<Video src={staticFile("video.mp4")} style={{width: "100%"}} />
<Audio src={staticFile("music.mp3")} volume={0.5} />
const {fontFamily} = loadFont({family: "Custom", url: staticFile("fonts/font.woff2")});
```

**Techniques:**

- staticFile("path"): Reference public/ assets; auto-encodes special characters
- Local assets: Place in public/; organize as images/, videos/, audio/, fonts/
- Remote URLs: Use directly without staticFile (CORS enabled)
- Google Fonts: `loadFont()` from @remotion/google-fonts/FontName
- Custom fonts: Load with @remotion/fonts; measure with measureText()
- Avoid: Native `<img>`, `<video>`, `<audio>` (don't wait for load)
