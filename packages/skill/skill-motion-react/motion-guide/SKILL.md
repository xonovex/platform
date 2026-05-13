---
name: motion-guide
description: "Use when adding or editing UI animations in React with the Motion library (formerly Framer Motion). Triggers on `motion/react` imports, `<motion.*>` components, `whileHover`, `whileTap`, `whileInView`, `useScroll`, `useSpring`, `layoutId`, `AnimatePresence`, and on prompts about entrance animations, hover effects, scroll reveals, layout transitions, or drag interactions, even when the user doesn't say 'Motion'. Skip CSS-only animations, GSAP, React Spring, and React Native Reanimated."
---

# Motion Animation Guidelines

## Requirements

- Install: `npm install motion`
- Import: `import { motion } from "motion/react"`

## Essentials

- **Entrance animations** - Use `initial`/`animate` with opacity+transform; 0.6-0.8s duration, ease `[0.22, 1, 0.36, 1]`, see [references/entrance.md](references/entrance.md)
- **Gesture interactions** - Use `whileHover`/`whileTap` with spring physics (stiffness: 300, damping: 20), see [references/gestures.md](references/gestures.md)
- **Scroll effects** - Use `whileInView` with `viewport={{once: true, amount: 0.3}}`; use `useScroll`/`useTransform` for parallax, see [references/scroll.md](references/scroll.md)
- **Layout animations** - Use `layout` prop for automatic FLIP; use `layoutId` for shared element morphing, see [references/layout.md](references/layout.md)
- **Stagger sequences** - Use variants with `staggerChildren: 0.1` and `delayChildren`, see [references/stagger.md](references/stagger.md)
- **Exit animations** - Wrap in `<AnimatePresence>` with `exit` prop, see [references/exit.md](references/exit.md)
- **Performance** - Only animate transform/opacity; use `useReducedMotion()` for accessibility, see [references/performance.md](references/performance.md)

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

## Gotchas

- `AnimatePresence` requires stable, unique `key` props on direct children — generated keys (`Math.random()`) recreate every render and break exit animations
- `initial={false}` on a first-render mounted component skips the initial animation — needed when restoring state from SSR or cache
- Layout animations (`layout` prop) trigger on every reflow — expensive components should also set `layoutDependency` to scope updates
- `useAnimate` and `useMotionValue` aren't React state — changing them doesn't re-render; reading them in JSX without a `MotionValue` consumer shows stale values

## Progressive Disclosure

- Read [references/entrance.md](references/entrance.md) - Load when creating fade-in, slide-up, or hero animations
- Read [references/gestures.md](references/gestures.md) - Load when adding hover, tap, or focus interactions
- Read [references/scroll.md](references/scroll.md) - Load when building scroll reveals or parallax effects
- Read [references/layout.md](references/layout.md) - Load when animating layout changes or shared elements
- Read [references/stagger.md](references/stagger.md) - Load when orchestrating sequential child animations
- Read [references/exit.md](references/exit.md) - Load when animating component unmounting
- Read [references/spring-physics.md](references/spring-physics.md) - Load when tuning spring stiffness, damping, mass
- Read [references/motion-values.md](references/motion-values.md) - Load when using useMotionValue, useSpring, useTransform, useAnimate, or MotionConfig
- Read [references/performance.md](references/performance.md) - Load when optimizing animations or supporting reduced motion
- Read [references/svg-path.md](references/svg-path.md) - Load when animating SVG paths or strokes
- Read [references/3d-effects.md](references/3d-effects.md) - Load when creating 3D cards, perspective, or rotations
- Read [references/text-effects.md](references/text-effects.md) - Load when animating text reveals, scramble, or typewriter effects
