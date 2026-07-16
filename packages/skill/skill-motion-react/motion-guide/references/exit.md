# exit: AnimatePresence Unmount Animations

Wrap conditionally rendered elements in `<AnimatePresence>` with an `exit` prop. Direct children need stable unique `key`s.

```tsx
function Modal({isOpen, onClose}) {
  return (
    <AnimatePresence>
      {isOpen && (
        <motion.div
          initial={{opacity: 0, scale: 0.95}}
          animate={{opacity: 1, scale: 1}}
          exit={{opacity: 0, scale: 0.95}}
          transition={{type: "spring", stiffness: 300}}>
          Modal content
        </motion.div>
      )}
    </AnimatePresence>
  );
}
```

- Exit modes (`mode` prop): `"sync"` (default, simultaneous), `"wait"` (exit before enter), `"popLayout"` (exit removed from flow)
- `onExitComplete={() => {...}}` fires when all exit animations finish
- Page transitions: `key={location.pathname}` triggers exit/enter on route change
- List collapse: `exit={{height: 0, opacity: 0}}` shrinks before removal
- Staggered exit: parent variant with `staggerDirection: -1` reverses order
