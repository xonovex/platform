# filename: 3d-effects

**Guideline:** Set `perspective` on parent; use `transformStyle: "preserve-3d"` on children; drive 3D via `useMotionValue` + `useSpring`.

**Rationale:** Perspective creates 3D depth; preserve-3d enables layering; motion values drive smooth tilt/flip without re-renders.

**Example:**

```tsx
function Card3D({children}: {children: React.ReactNode}) {
  const x = useMotionValue(0);
  const y = useMotionValue(0);
  const rotateX = useSpring(useTransform(y, [-0.5, 0.5], [15, -15]));
  const rotateY = useSpring(useTransform(x, [-0.5, 0.5], [-15, 15]));

  return (
    <motion.div
      style={{rotateX, rotateY, transformStyle: "preserve-3d"}}
      onMouseMove={(e) => {
        const rect = e.currentTarget.getBoundingClientRect();
        x.set((e.clientX - rect.left) / rect.width - 0.5);
      }}>
      {children}
    </motion.div>
  );
}
```

**Techniques:**

- Parent perspective: `style={{perspective: 1000}}`; higher = less depth distortion
- Flip cards: `animate={{rotateY: isFlipped ? 180 : 0}}`; use `backfaceVisibility: "hidden"`
- 3D carousel: Position items with `rotateY(angle)` and `translateZ()` in transforms
- Layered stacks: `translateZ(i * 20px)` for depth; `whileHover={{rotateX, rotateY}}` on container
- Shine effect: Use `useTransform()` to derive shine position from mouse coordinates; gradient follows
- Transform properties: `rotateX/Y/Z`, `x/y/z`, `scale/scaleX/Y/Z`, `transformOrigin`
