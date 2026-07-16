# scroll: Reveals and Parallax

Use `whileInView` for scroll-triggered reveals; use `useScroll()` + `useTransform()` for scroll-linked parallax.

```tsx
function ScrollReveal({children}) {
  return (
    <motion.div
      initial={{opacity: 0, y: 50}}
      whileInView={{opacity: 1, y: 0}}
      viewport={{once: true, amount: 0.3}}
      transition={{duration: 0.6}}>
      {children}
    </motion.div>
  );
}

// Parallax
function Parallax() {
  const ref = useRef(null);
  const {scrollYProgress} = useScroll({
    target: ref,
    offset: ["start end", "end start"],
  });
  const backgroundY = useTransform(scrollYProgress, [0, 1], [0, -150]);
  return <motion.div ref={ref} style={{y: backgroundY}} />;
}
```

- `viewport`: `amount` (0-1 visible fraction), `margin` (CSS px), `once: true` (single fire)
- `useScroll()`: `scrollX/Y` + `scrollXProgress/scrollYProgress` (0-1)
- `useInView(ref, {once: true})`: imperative boolean hook
- Progress bar: `scaleX: scrollYProgress` with `transformOrigin: "left"`
- Staggered reveal: `variants` + `staggerChildren` inside `whileInView`
