# text: Text and Typography

**Guideline:** Load fonts first; use measureText() for dimensions; fitText() to auto-scale; animate with frame-based calculations.

**Rationale:** Font loading is async; measuring text enables precise layout; frame-based animation ensures consistent timing.

**Example:**
```tsx
import {loadFont} from "@remotion/google-fonts/Roboto";
import {measureText, fitText} from "@remotion/layout-utils";

const {fontFamily} = loadFont(); // Google Font
const {width} = measureText({text: "Hello", fontFamily, fontSize: 48});
const {fontSize} = fitText({text, withinWidth: 500, fontFamily});
```

**Techniques:**
- Google Fonts: `loadFont()` from @remotion/google-fonts/FontName
- Local fonts: `loadFont({family, url: staticFile(), weight})`
- measureText(): Returns {width, height}; use validateFontIsLoaded: true
- fitText(): Auto-scale fontSize to fit container width
- Typewriter: `Math.floor(frame / fps * charsPerSecond)` to slice text
- Word highlight: spring() to animate background scaleX across word
- Stagger reveal: Delay each element by `index * delaySeconds * fps`
