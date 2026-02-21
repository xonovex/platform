# performance: Performance and Accessibility

**Guideline:** Animate only GPU-accelerated transforms (x/y/scale/rotate) and opacity; use scaleX/Y or layout for size changes; support prefers-reduced-motion.

**Rationale:** Transform/opacity don't trigger layout recalculation; width/height cause jank. Respecting prefers-reduced-motion is legal requirement (WCAG 2.1).

**Example:**

```tsx
import {motion, useReducedMotion} from "motion/react";

function AccessibleDiv({children}) {
  const shouldReduceMotion = useReducedMotion();
  return (
    <motion.div
      initial={shouldReduceMotion ? false : {opacity: 0, y: 20}}
      animate={{opacity: 1, y: 0}}
      transition={shouldReduceMotion ? {duration: 0} : {duration: 0.6}}>
      {children}
    </motion.div>
  );
}
```

**Techniques:**

- GPU-accelerated: x, y, z, scale, rotate, opacity, skew (no layout recalc)
- Never animate: width, height, margin, padding, left/top/right/bottom (triggers jank)
- useReducedMotion(): Check for prefers-reduced-motion; skip animations if true
- Layout prop: Use `layout` for size/position changes instead of animating dimensions
- LazyMotion: Wrap app with LazyMotion + domAnimation to reduce bundle (use `m` instead of `motion`)
- willChange: Add for high-frequency animations
- Accessible wrapper: Create MotionSafe component that skips animations for reduced-motion users
