# lottie: Lottie Animations

**Guideline:** Use `delayRender()` pattern to load Lottie JSON before rendering compositions.

**Rationale:** Animations must load asynchronously. `delayRender()` pauses frame rendering until `continueRender()` is called after JSON loads.

**Example:**

```tsx
import {Lottie, useVideoConfig} from "@remotion/lottie";
import {continueRender, delayRender} from "remotion";

const handle = delayRender();

export const MyAnimation = () => {
  const [animData, setAnimData] = useState<LottieAnimationData | null>(null);
  const {fps} = useVideoConfig();

  useEffect(() => {
    fetch("https://lottiefiles.com/animations/123.json")
      .then((r) => r.json())
      .then((data) => {
        setAnimData(data);
        continueRender(handle);
      })
      .catch((err) => cancelRender(err));
  }, [handle]);

  return animData ? <Lottie animationData={animData} /> : null;
};
```

**Techniques:**

- Install `@remotion/lottie` package before use
- Call `delayRender()` to pause frame rendering while loading
- Fetch animation JSON from remote (Lottie Files) or local with `staticFile()`
- Store fetched JSON in state with `useState<LottieAnimationData | null>()`
- Call `continueRender(handle)` when JSON loaded successfully
- Call `cancelRender(err)` if fetch fails to report error
- Return null while loading to prevent rendering with undefined data
- Pass animation data to `<Lottie animationData={...} />` component
- Apply width/height styles via style props
