# svg-path: Drawing Strokes with pathLength

Animate `pathLength` from 0 to 1 on `motion.path` for drawing effects (Motion computes the stroke-dash math); pair with `opacity` for simultaneous fade.

```tsx
function DrawingIcon() {
  return (
    <svg width="100" height="100" viewBox="0 0 100 100">
      <motion.path
        d="M 10 50 Q 50 10 90 50 Q 50 90 10 50"
        fill="none"
        stroke="#667eea"
        strokeWidth={2}
        initial={{pathLength: 0, opacity: 0}}
        animate={{pathLength: 1, opacity: 1}}
        transition={{duration: 1.5}}
      />
    </svg>
  );
}
```

- Progress circles: `animate={{pathLength: progress}}` (0 to dynamic value)
- Staggered paths: `variants` + `custom` prop for per-path delay
- Flowing lines: `strokeDasharray="10 5"` + `strokeDashoffset: [0, -30]` with `repeat: Infinity`
- Stroke caps: `strokeLinecap="round"` + `strokeLinejoin="round"`
