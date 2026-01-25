# filename: scroll

**Guideline:** Use `whileInView` for scroll-triggered reveals; use `useScroll()` + `useTransform()` for precise parallax and scroll-linked effects.

**Rationale:** `whileInView` leverages Intersection Observer; `useScroll()` enables smooth parallax without scroll listeners; `useTransform` maps scroll range to outputs.

**Example:**
```tsx
function ScrollReveal({children}) {
  return (
    <motion.div
      initial={{opacity: 0, y: 50}}
      whileInView={{opacity: 1, y: 0}}
      viewport={{once: true, amount: 0.3}}
      transition={{duration: 0.6}}
    >
      {children}
    </motion.div>
  );
}

// Parallax
function Parallax() {
  const ref = useRef(null);
  const {scrollYProgress} = useScroll({target: ref, offset: ["start end", "end start"]});
  const backgroundY = useTransform(scrollYProgress, [0, 1], [0, -150]);
  return <motion.div ref={ref} style={{y: backgroundY}} />;
}
```

**Techniques:**
- `whileInView`: Animates when element enters viewport; `viewport={{once: true, amount: 0.3}}` triggers at 30% visible
- `useScroll()`: Returns `scrollX/Y` and `scrollXProgress/scrollYProgress` (0-1 normalized)
- `useTransform(scrollProgress, [0, 1], [outputStart, outputEnd])`: Map scroll to colors, positions, scale
- `useInView(ref, {once: true})`: Imperative boolean hook for manual control
- Staggered scroll reveal: `variants` with `staggerChildren` inside `whileInView`
- Scroll progress bar: `scaleX: scrollYProgress` with `transformOrigin: "left"`
- Viewport options: `amount` (0-1), `margin` (CSS pixels), `once: true` (single animation)
