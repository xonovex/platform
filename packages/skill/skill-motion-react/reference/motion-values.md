# filename: motion-values

**Guideline:** Use `useMotionValue` for high-frequency updates (mouse, scroll) without re-renders; derive with `useTransform`; smooth with `useSpring`.

**Rationale:** Motion values bypass React rendering; enables 60fps tracking; perfect for animations that don't need state.

**Example:**

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

**Techniques:**

- `useMotionValue(initial)`: Create raw value; `.set()` updates without renders
- `useTransform(value, [input], [output])`: Map range; supports colors, strings, numbers
- `useSpring(value, config)`: Add spring physics (stiffness, damping, mass)
- `useMotionTemplate`: Dynamic CSS strings with backticks: `` `radial-gradient(... ${x}px ...)` ``
- `useScroll()`: Returns `scrollX`/`scrollY`/`scrollXProgress`/`scrollYProgress` (0-1)
- Subscription: `value.on("change", latest => {...})`; unsubscribe in cleanup
- `useAnimate()`: Imperative control - `const [scope, animate] = useAnimate()`, then `await animate(scope.current, {...})`
- Performance: Never use state for real-time values; motion values skip renders for 60fps tracking
