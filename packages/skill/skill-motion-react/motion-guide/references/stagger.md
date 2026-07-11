# stagger: Sequential Child Animations

Use `staggerChildren` in a parent variant's `transition` to sequence children; never use manual `delay: i * 0.1`. Give list items unique `key`s.

```tsx
const container = {
  hidden: {},
  visible: {
    transition: {staggerChildren: 0.1, delayChildren: 0.2},
  },
};
const item = {
  hidden: {opacity: 0, y: 20},
  visible: {opacity: 1, y: 0, transition: {duration: 0.5}},
};

<motion.ul variants={container} initial="hidden" animate="visible">
  {items.map((i, idx) => (
    <motion.li key={idx} variants={item}>
      {i}
    </motion.li>
  ))}
</motion.ul>;
```

- `staggerChildren` = per-child delay; `delayChildren` = delay before first
- Timing: grids 0.02-0.05, lists 0.08-0.12, hero sections 0.15-0.2
- Reverse exit: `exit: {transition: {staggerDirection: -1}}`
- Scroll-triggered: `whileInView="visible"` + `viewport={{once: true}}`
