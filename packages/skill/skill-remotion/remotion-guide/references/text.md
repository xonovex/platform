# text: Text and Typography

Load fonts before layout: `loadFont()` from `@remotion/google-fonts/FontName`, or `loadFont({family, url: staticFile(), weight})` for local fonts. Measure with `@remotion/layout-utils`: `measureText({text, fontFamily, fontSize})` → `{width, height}` (pass `validateFontIsLoaded: true`); `fitText({text, withinWidth, fontFamily})` → `{fontSize}` to auto-scale.

```tsx
import {loadFont} from "@remotion/google-fonts/Roboto";
import {fitText, measureText} from "@remotion/layout-utils";

const {fontFamily} = loadFont();
const {width} = measureText({text: "Hello", fontFamily, fontSize: 48});
const {fontSize} = fitText({text, withinWidth: 500, fontFamily});
```

- Typewriter: slice text by `Math.floor((frame / fps) * charsPerSecond)`
