# filename: entrance

**Guideline:** Use `initial`/`animate` with opacity and transforms (y, x, scale); apply Apple easing `[0.22, 1, 0.36, 1]` for smooth deceleration.

**Rationale:** Opacity + transforms are GPU-optimized; transforms avoid layout thrashing; easing creates polished, professional feel.

**Example:**
```tsx
function FadeUp({children}) {
  return (
    <motion.div
      initial={{opacity: 0, y: 20}}
      animate={{opacity: 1, y: 0}}
      transition={{duration: 0.6, ease: [0.22, 1, 0.36, 1]}}
    >
      {children}
    </motion.div>
  );
}

// Staggered hero
<motion.section initial="hidden" animate="visible"
  variants={{visible: {transition: {staggerChildren: 0.15}}}}>
  <motion.h1 variants={{hidden: {opacity: 0, y: 30}, visible: {opacity: 1, y: 0}}} />
</motion.section>
```

**Techniques:**
- Fade up: `initial={{opacity: 0, y: 20}}` → `animate={{opacity: 1, y: 0}}`
- Fade in scale: `initial={{opacity: 0, scale: 0.95}}` for zoom-in effect
- Slide directions: `x: -50` (left), `x: 50` (right), `y: 50` (bottom)
- Spring entrance: `type: "spring", stiffness: 300, damping: 20` for bouncy
- Duration hierarchy: Headlines 0.8s, subheadings 0.6s, body 0.5s, small 0.3-0.4s
- Stagger children: Parent variant with `staggerChildren: 0.1-0.15` for cascading
- Apple easing: `ease: [0.22, 1, 0.36, 1]` for professional smooth deceleration
- Always combine opacity with transform for optimal performance; never use margin for animation
