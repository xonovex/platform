---
name: motion-guidelines
description: >-
  Trigger on Motion animations, `whileHover`, `whileInView`, scroll effects, spring physics, gesture animations. Use when creating React animations with Motion. Apply for entrance animations, hover effects, scroll reveals, layout transitions, drag interactions. Keywords: Motion, motion/react, animation, React, spring physics, whileHover, whileTap, useScroll, layoutId, AnimatePresence.
---

# Motion Animation Guidelines

## Requirements

- Install: `npm install motion`
- Import: `import { motion } from "motion/react"`

## Essentials

- **Entrance animations** - Use `initial`/`animate` with opacity+transform; 0.6-0.8s duration, ease `[0.22, 1, 0.36, 1]`, see [reference/entrance.md](reference/entrance.md)
- **Gesture interactions** - Use `whileHover`/`whileTap` with spring physics (stiffness: 300, damping: 20), see [reference/gestures.md](reference/gestures.md)
- **Scroll effects** - Use `whileInView` with `viewport={{once: true, amount: 0.3}}`; use `useScroll`/`useTransform` for parallax, see [reference/scroll.md](reference/scroll.md)
- **Layout animations** - Use `layout` prop for automatic FLIP; use `layoutId` for shared element morphing, see [reference/layout.md](reference/layout.md)
- **Stagger sequences** - Use variants with `staggerChildren: 0.1` and `delayChildren`, see [reference/stagger.md](reference/stagger.md)
- **Exit animations** - Wrap in `<AnimatePresence>` with `exit` prop, see [reference/exit.md](reference/exit.md)
- **Performance** - Only animate transform/opacity; use `useReducedMotion()` for accessibility, see [reference/performance.md](reference/performance.md)

## Example

```tsx
import {motion} from "motion/react";

export function FadeUp({children}: {children: React.ReactNode}) {
  return (
    <motion.div
      initial={{opacity: 0, y: 20}}
      animate={{opacity: 1, y: 0}}
      transition={{duration: 0.6, ease: [0.22, 1, 0.36, 1]}}>
      {children}
    </motion.div>
  );
}

export function HoverCard({children}: {children: React.ReactNode}) {
  return (
    <motion.div
      whileHover={{y: -8, boxShadow: "0 20px 40px rgba(0,0,0,0.12)"}}
      transition={{type: "spring", stiffness: 300, damping: 20}}>
      {children}
    </motion.div>
  );
}
```

## Progressive Disclosure

- Read [reference/entrance.md](reference/entrance.md) - When creating fade-in, slide-up, or hero animations
- Read [reference/gestures.md](reference/gestures.md) - When adding hover, tap, or focus interactions
- Read [reference/scroll.md](reference/scroll.md) - When building scroll reveals or parallax effects
- Read [reference/layout.md](reference/layout.md) - When animating layout changes or shared elements
- Read [reference/stagger.md](reference/stagger.md) - When orchestrating sequential child animations
- Read [reference/exit.md](reference/exit.md) - When animating component unmounting
- Read [reference/spring-physics.md](reference/spring-physics.md) - When tuning spring stiffness, damping, mass
- Read [reference/motion-values.md](reference/motion-values.md) - When using useMotionValue, useSpring, useTransform, useAnimate, or MotionConfig
- Read [reference/performance.md](reference/performance.md) - When optimizing animations or supporting reduced motion
- Read [reference/svg-path.md](reference/svg-path.md) - When animating SVG paths or strokes
- Read [reference/3d-effects.md](reference/3d-effects.md) - When creating 3D cards, perspective, or rotations
- Read [reference/text-effects.md](reference/text-effects.md) - When animating text reveals, scramble, or typewriter effects
