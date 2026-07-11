# performance: GPU Transforms and Reduced Motion

Animate only `x/y/z`, `scale`, `rotate`, `skew`, `opacity` (no layout recalc); use `scaleX/Y` or the `layout` prop for size changes. Gate animations on `useReducedMotion()` (WCAG 2.1 prefers-reduced-motion).

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

- Never animate `width`, `height`, `margin`, `padding`, `left/top/right/bottom` (jank)
- `LazyMotion` + `domAnimation` shrinks bundle — use `m` instead of `motion`
- Add `willChange` for high-frequency animations
