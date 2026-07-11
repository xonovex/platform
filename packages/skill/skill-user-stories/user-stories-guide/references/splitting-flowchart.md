# user-stories: Story-Splitting Flowchart and Vertical Slicing

The fuller decision tree for breaking a too-big story into smaller ones — a superset of SPIDR (see [splitting-spidr.md](splitting-spidr.md)): confirm the item is a valuable story that is too big, apply a pattern, then evaluate the split.

## Preconditions (before you split)

1. **Valuable story, not a task/component?** "Build the API" is a task — do not "split" it; restructure into vertical stories.
2. **Too big?** Only split a story that fails INVEST **Small** (see [invest-and-smart.md](invest-and-smart.md)). One that fits a sprint, leave alone.

## The splitting patterns

- **Workflow Steps** — deliver the core step of a multi-step workflow first.
- **Operations (CRUD)** — create first, then read/update/delete.
- **Business Rule Variations** — common rule first.
- **Variations in Data** — one region/format/currency first.
- **Data Entry Methods** — simplest input method first.
- **Major Effort** — do the costly first case first; later cases get cheaper.
- **Simple / Complex** — pull out the simple core, defer complex cases.
- **Deferred Performance (and other NFRs)** — make it work first, fast/secure/scalable later as explicit follow-up stories.
- **Break Out a Spike** — last resort: research for knowledge, not value.

## The two evaluation rules

1. **Prefer splits that enable deprioritization** — you can ship high-value slices and drop low-value ones. If every slice is needed for any value, the split bought nothing.
2. **Favor equal-sized splits** — several roughly equal stories over one big + one tiny.

## Vertical slicing and the walking skeleton

Every story is a **vertical slice**: a thin path through all layers it needs (UI, logic, data) delivering usable value, never a horizontal layer. The first slice of a new capability is the **walking skeleton**: a tiny end-to-end implementation linking the main architectural components. It is **production code with tests** — not a throwaway prototype and not a research spike.

## Example

```
Story 1 (walking skeleton): request → match → charge saved card → show "Paid"
Story 2 (Variations in Data):  multi-currency charging
Story 3 (Business Rule):       split-fare
Story 4 (Deferred NFR):        sub-200ms charge confirmation
```

Stories 2–4 are independently shippable and can be reordered or dropped by value.

## Cross-references

- SPIDR, the compact starter set this flowchart subsumes — [splitting-spidr.md](splitting-spidr.md).
- INVEST, the gate a story must fail (Small) before you split — [invest-and-smart.md](invest-and-smart.md).
- The FDD feature list, a different decomposition lens — **fdd-guide**.
