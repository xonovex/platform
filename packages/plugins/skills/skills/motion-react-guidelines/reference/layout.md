# filename: layout

**Guideline:** Use `layout` prop for automatic FLIP animations on size/position changes; use `layoutId` for shared element transitions between DOM nodes.

**Rationale:** `layout` auto-animates position/size without manually specifying targets; `layoutId` enables seamless transitions across screens (gallery → modal).

**Example:**
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

// Shared element transition
<motion.div layoutId={`card-${id}`} onClick={() => setSelected(item)}>
  <motion.img layoutId={`image-${id}`} src={item.image} />
</motion.div>

{selected && (
  <motion.div layoutId={`card-${selected.id}`} style={{position: "fixed"}}>
    <motion.img layoutId={`image-${selected.id}`} src={selected.image} />
  </motion.div>
)}
```

**Techniques:**
- `layout`: Animate position and size changes automatically (FLIP algorithm)
- `layout="position"`: Only position (for text); `layout="size"`: Only size
- `layoutId`: Same ID on elements across different DOM states for smooth morphing
- `<Reorder.Group values={items} onReorder={setItems}>`: Drag-to-reorder with auto-layout
- `<Reorder.Item value={item}>`: Draggable list item
- Accordion: Combine `layout` with `AnimatePresence` for smooth expand/collapse
- Tabs indicator: `layoutId="tab-indicator"` + conditional render for smooth underline transition
- Always wrap conditional elements in `<AnimatePresence>` for exit animations
