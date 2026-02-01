# filename: svg-path

**Guideline:** Animate SVG paths with `pathLength` from 0 to 1 for drawing effects; pair with `opacity` for simultaneous fade-in.

**Rationale:** Motion's `pathLength` auto-calculates stroke-dash math; simpler than manual `strokeDasharray`/`strokeDashoffset`.

**Example:**

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

**Techniques:**

- Path animation: `initial={{pathLength: 0}}` → `animate={{pathLength: 1}}`; 0-1 range
- Staggered paths: Use `variants` + `custom` prop for per-path delays
- Progress circles: `animate={{pathLength: progress}}` (0 to dynamic value)
- Checkmarks: Combine `pathLength` animation with `opacity` transition on different durations
- Flowing lines: Use `strokeDasharray="10 5"` + `strokeDashoffset: [0, -30]` with `repeat: Infinity`
- Connection lines: Bezier curves `C` command for smooth connections; `rotateY` adjustment for vertical lines
- Stroke caps: `strokeLinecap="round"` + `strokeLinejoin="round"` for smoother endpoints
