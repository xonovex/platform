# three-d: 3D Content with Three.js

**Guideline:** Use ThreeCanvas (not react-three-fiber Canvas); animate with useCurrentFrame, not useFrame(); use layout="none" in Sequences.

**Rationale:** useFrame() updates real-time, not frame-perfect; Remotion requires frame-based deterministic animation.

**Example:**
```tsx
import {ThreeCanvas} from "@remotion/three";
import {useCurrentFrame, useVideoConfig} from "remotion";

function RotatingCube() {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  const rotation = (frame / (2 * fps)) * Math.PI * 2;
  return (
    <mesh rotation={[0, rotation, 0]}>
      <boxGeometry args={[1, 1, 1]} />
      <meshStandardMaterial color="orange" />
    </mesh>
  );
}
```

**Techniques:**
- ThreeCanvas: Remotion-specific wrapper (not R3F Canvas)
- useCurrentFrame() + fps: Calculate frame-based transforms (rotation, position)
- interpolate(): Linear motion and camera movement
- spring(): Bouncing and physics-based animation with config {damping, stiffness}
- PerspectiveCamera: Animate position/lookAt with useCurrentFrame
- Sequence layout="none": Required for ThreeCanvas (default layout breaks 3D)
- Lighting: ambientLight (base), directionalLight (shadows), pointLight (highlights)
- Stagger: Use `frame - (index * fps * delay)` for sequenced objects
