# spring-physics: Tuning stiffness, damping, mass

Use `type: "spring"` for interactive/interruptible motion (gestures, drag, layout); use `duration` for synchronized sequences (entrances, scroll reveals, page transitions).

```tsx
// Snappy (buttons, quick interactions)
<motion.div
  whileHover={{scale: 1.05}}
  transition={{type: "spring", stiffness: 400, damping: 25}}
/>

// Bouncy (playful entrance)
<motion.div
  initial={{scale: 0.9}}
  animate={{scale: 1}}
  transition={{type: "spring", stiffness: 200, damping: 8}}
/>
```

- `stiffness: 100-500`: response speed; higher = faster snap
- `damping: 10-50`: overshoot; higher = less bounce
- `mass: 0.5-3`: perceived weight; higher = heavier/slower
- Presets: snappy 400/25 · smooth 300/20 · bouncy 200/8 · gentle 150/15 · heavy 100/15 mass 2
- `useSpring()` smooths raw motion values (mouse tracking, scroll)
