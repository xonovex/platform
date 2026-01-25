# filename: text-effects

**Guideline:** Split text into words/characters in spans; use `overflow: hidden` with `y: ["100%", 0]` for elegant reveal animations.

**Rationale:** Splitting enables per-element stagger; `overflow: hidden` masks overflow for clean Y-translate reveals; avoid animating entire text blocks.

**Example:**
```tsx
function TextReveal({text}) {
  return (
    <motion.div initial="hidden" animate="visible" variants={{
      hidden: {},
      visible: {transition: {staggerChildren: 0.1}}
    }}>
      {text.split(" ").map((word, i) => (
        <span key={i} style={{display: "inline-block", overflow: "hidden"}}>
          <motion.span variants={{
            hidden: {y: "100%", opacity: 0},
            visible: {y: 0, opacity: 1}
          }} style={{display: "inline-block"}}>
            {word}
          </motion.span>
          <span>&nbsp;</span>
        </span>
      ))}
    </motion.div>
  );
}
```

**Techniques:**
- Word reveal: Split by space, `y: ["100%", 0]`, stagger 0.05-0.15
- Character reveal: Split by `""`, stagger 0.02-0.05, add `y: 20` for bounce effect
- Typewriter: Use `useState` + `setInterval`, slice() text incrementally, add blinking cursor
- Wavy text: Per-character `y: [0, -10, 0]` with delay based on index
- Gradient animation: `backgroundPosition` with `backgroundSize: "300% 100%"`
- Word highlight: Overlay `scaleX: 0` → `scaleX: 1` highlight box behind word
- Overflow masking: Always wrap in `<span style={{overflow: "hidden"}}>` for clean reveals
- Keep stagger delays small: 0.02-0.1 for natural flow
