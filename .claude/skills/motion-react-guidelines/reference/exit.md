# filename: exit

**Guideline:** Wrap conditionally rendered elements in `<AnimatePresence>` to enable exit animations; React removes elements instantly without it.

**Rationale:** AnimatePresence delays DOM removal until exit animation completes; enables smooth transitions and proper cleanup.

**Example:**
```tsx
function Modal({isOpen, onClose}) {
  return (
    <AnimatePresence>
      {isOpen && (
        <motion.div
          initial={{opacity: 0, scale: 0.95}}
          animate={{opacity: 1, scale: 1}}
          exit={{opacity: 0, scale: 0.95}}
          transition={{type: "spring", stiffness: 300}}
        >
          Modal content
        </motion.div>
      )}
    </AnimatePresence>
  );
}
```

**Techniques:**
- Always wrap conditionals: `<AnimatePresence>{condition && <motion.div exit={{...}} />}</AnimatePresence>`
- Unique keys: Required for list animations; prevents animation identity loss on reorder
- Exit modes: `"sync"` (enter/exit simultaneous), `"wait"` (exit first, then enter), `"popLayout"` (exit removed from flow)
- Exit callbacks: `onExitComplete={() => {...}}` fires when all exit animations finish
- Page transitions: Use `key={location.pathname}` with router to trigger exit/enter on route change
- List animations: `height: 0` on exit shrinks before removal; `opacity: 0` fades
- Staggered exit: Use parent variant with `staggerDirection: -1` to reverse animation order
- Modal backdrop: Animate both backdrop (fade) and content (scale) simultaneously for depth
