# layout: FLIP and Shared-Element Transitions

Add the `layout` prop for automatic FLIP animation of size/position changes; use matching `layoutId` on elements across different DOM states for shared-element morphing.

```tsx
function ExpandingCard({expanded}) {
  return (
    <motion.div
      layout
      style={{width: expanded ? 400 : 200}}
      transition={{type: "spring", stiffness: 300}}
    />
  );
}

// Shared element transition (gallery → modal)
<motion.div layoutId={`card-${id}`} onClick={() => setSelected(item)}>
  <motion.img layoutId={`image-${id}`} src={item.image} />
</motion.div>;

{
  selected && (
    <motion.div layoutId={`card-${selected.id}`} style={{position: "fixed"}}>
      <motion.img layoutId={`image-${selected.id}`} src={selected.image} />
    </motion.div>
  );
}
```

- `layout="position"` (text only) / `layout="size"` (size only)
- `<Reorder.Group values={items} onReorder={setItems}>` + `<Reorder.Item value={item}>`: drag-to-reorder with auto-layout
- Accordion: combine `layout` with `AnimatePresence` for expand/collapse
- Tabs indicator: shared `layoutId="tab-indicator"` for sliding underline
