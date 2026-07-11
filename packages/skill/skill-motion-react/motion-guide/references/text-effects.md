# text-effects: Word and Character Reveals

Split text into per-word/char spans for stagger; wrap each in `overflow: hidden` and translate `y: ["100%", 0]` for masked reveals.

```tsx
function TextReveal({text}) {
  return (
    <motion.div
      initial="hidden"
      animate="visible"
      variants={{
        hidden: {},
        visible: {transition: {staggerChildren: 0.1}},
      }}>
      {text.split(" ").map((word, i) => (
        <span key={i} style={{display: "inline-block", overflow: "hidden"}}>
          <motion.span
            variants={{
              hidden: {y: "100%", opacity: 0},
              visible: {y: 0, opacity: 1},
            }}
            style={{display: "inline-block"}}>
            {word}
          </motion.span>
          <span>&nbsp;</span>
        </span>
      ))}
    </motion.div>
  );
}
```

- Word reveal: split by `" "`, stagger 0.05-0.15
- Character reveal: split by `""`, stagger 0.02-0.05, add `y: 20` for bounce
- Typewriter: `useState` + `setInterval`, `slice()` incrementally, blinking cursor
- Wavy: per-char `y: [0, -10, 0]` with index-based delay
- Gradient sweep: animate `backgroundPosition` with `backgroundSize: "300% 100%"`
