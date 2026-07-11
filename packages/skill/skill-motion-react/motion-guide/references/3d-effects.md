# 3d-effects: 3D Cards and Perspective

Set `perspective` on the parent, `transformStyle: "preserve-3d"` on children; drive tilt/flip via `useMotionValue` + `useSpring` (no re-renders).

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

- Parent perspective: `style={{perspective: 1000}}` — higher = less depth distortion
- Flip cards: `animate={{rotateY: isFlipped ? 180 : 0}}` + `backfaceVisibility: "hidden"`
- 3D carousel: position items with `rotateY(angle)` + `translateZ()`
- Layered stacks: `translateZ(i * 20px)` for depth
- Shine: `useTransform()` to derive gradient position from mouse coords
- Transform props: `rotateX/Y/Z`, `x/y/z`, `scale/scaleX/Y/Z`, `transformOrigin`
