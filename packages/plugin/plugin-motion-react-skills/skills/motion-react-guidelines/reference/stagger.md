# filename: stagger

**Guideline:** Use `staggerChildren` in parent variant transitions to orchestrate sequential child animations; never use manual delays.

**Rationale:** Variants + stagger maintains maintainability; auto-sequencing prevents timing bugs; reversible with `staggerDirection: -1`.

**Example:**

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

**Techniques:**

- Parent variant: `staggerChildren` (per-child delay), `delayChildren` (initial delay before first)
- Child variant: individual `initial`/`visible`/`exit` states
- Grid stagger: 0.02-0.05 for dense layouts; 0.08-0.12 for lists; 0.15-0.2 for hero sections
- Reverse exit: `exit: {transition: {staggerDirection: -1}}` for backwards animation
- Scroll trigger: `whileInView="visible"` + `viewport={{once: true}}` for scroll-triggered stagger
- Always use unique `key` on list items for proper animation tracking
- Never use `delay: i * 0.1`; always use variants + staggerChildren for maintainability
