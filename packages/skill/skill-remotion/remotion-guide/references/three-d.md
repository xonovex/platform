# three-d: 3D Content with Three.js

Use `ThreeCanvas` from `@remotion/three` (not react-three-fiber's `Canvas`) and animate with `useCurrentFrame()` — never `useFrame()`, which updates in real time and breaks deterministic rendering. Wrap in a `<Sequence layout="none">` (default layout breaks 3D).

```tsx
import {ThreeCanvas} from "@remotion/three";

const frame = useCurrentFrame();
const {fps} = useVideoConfig();
const rotation = (frame / (2 * fps)) * Math.PI * 2;
return (
  <mesh rotation={[0, rotation, 0]}>
    <boxGeometry args={[1, 1, 1]} />
    <meshStandardMaterial color="orange" />
  </mesh>
);
```
