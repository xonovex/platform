# motion-values: useMotionValue, useTransform, useSpring, useAnimate

Use `useMotionValue` for high-frequency updates (mouse, scroll) that bypass React rendering; derive with `useTransform`; smooth with `useSpring`.

```tsx
function MouseTracker() {
  const x = useMotionValue(0);
  const springX = useSpring(x, {stiffness: 150, damping: 15});

  return (
    <div onMouseMove={(e) => x.set(e.clientX)}>
      <motion.div style={{x: springX}} />
    </div>
  );
}

// Derive multiple values from one source
const background = useTransform(x, [-100, 0, 100], ["red", "white", "green"]);
const scale = useTransform(x, [-100, 100], [0.5, 1.5]);
```

- `useMotionValue(initial)`: `.set()` updates without re-render
- `useTransform(value, [input], [output])`: maps colors, strings, numbers
- `useSpring(value, config)`: spring physics (stiffness, damping, mass)
- `useMotionTemplate`: dynamic CSS string, e.g. `` `radial-gradient(... ${x}px ...)` ``
- `useScroll()`: returns `scrollX`/`scrollY`/`scrollXProgress`/`scrollYProgress` (0-1)
- Subscribe: `value.on("change", latest => {...})`; unsubscribe in cleanup
- `useAnimate()`: imperative — `const [scope, animate] = useAnimate()`, then `await animate(scope.current, {...})`
