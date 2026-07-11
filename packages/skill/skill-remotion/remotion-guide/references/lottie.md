# lottie: Lottie Animations

Load Lottie JSON async, gating render with `delayRender()`/`continueRender()`. Call `continueRender(handle)` once JSON is in state; `cancelRender(err)` on fetch failure; return `null` while loading.

```tsx
import {Lottie} from "@remotion/lottie";
import {cancelRender, continueRender, delayRender} from "remotion";

const handle = delayRender();
useEffect(() => {
  fetch(url)
    .then((r) => r.json())
    .then((data) => {
      setAnimData(data);
      continueRender(handle);
    })
    .catch((err) => cancelRender(err));
}, [handle]);
return animData ? <Lottie animationData={animData} /> : null;
```
