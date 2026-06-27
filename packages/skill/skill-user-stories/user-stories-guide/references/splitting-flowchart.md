# user-stories: Story-Splitting Flowchart and Vertical Slicing

The splitting flowchart is the fuller decision tree for breaking a too-big story into smaller ones. It is a superset of SPIDR (see [splitting-spidr.md](splitting-spidr.md)): first confirm the item is actually a valuable story that is too big, then apply a splitting pattern, then evaluate the resulting split. Throughout, every slice is **vertical** — a thin end-to-end path, never a horizontal layer.

## Preconditions (before you split)

1. **Is it a valuable story, not a task or component?** A story delivers user value end to end. "Build the API" is a task/component, not a story — do not "split" it; restructure the work into vertical stories.
2. **Is it too big?** Only split a story that fails INVEST's **Small** (see [invest-and-smart.md](invest-and-smart.md)). A story that already fits a sprint should be left alone.

If it is a valuable story and too big, choose a splitting pattern.

## The splitting patterns

- **Workflow Steps** — split off the steps of a multi-step workflow; deliver the core step first.
- **Operations (CRUD)** — split by operation: create first, then read, update, delete.
- **Business Rule Variations** — split off variations of a business rule; the common rule first.
- **Variations in Data** — split by data subset: one region, format, or currency first.
- **Data Entry Methods** — split by how data is entered: the simplest input method first.
- **Major Effort** — when most of the cost is in the first case, do that first; later cases get cheaper.
- **Simple / Complex** — pull out the simple core and defer the complex cases as their own stories.
- **Deferred Performance (and other NFRs)** — make it work first, fast/secure/scalable later (as explicit follow-up stories).
- **Break Out a Spike** — last resort: when you cannot see how to split, do research first to gain knowledge (knowledge, not value).

## The two evaluation rules

After splitting, judge the result against two rules:

1. **Prefer splits that enable deprioritization** — a good split lets you ship the high-value slices and deprioritize or discard the low-value ones. If you must build every slice for any value, the split bought you nothing.
2. **Favor equal-sized splits** — prefer several small, roughly equal stories over one big story plus one tiny story. Equal sizes flow and estimate better.

## Vertical slicing and the walking skeleton

Every story is a **vertical slice**: a thin path through all the layers it needs (UI, logic, data) that delivers usable value, rather than a horizontal layer that delivers none on its own.

The first vertical slice of a new capability is the **walking skeleton**: a tiny implementation that performs a small end-to-end function. It need not use the final architecture, but it should link together the main architectural components so that architecture and functionality can evolve in parallel. Crucially, a walking skeleton is **production code with tests** — not a throwaway prototype and not a research spike.

## BAD → GOOD

BAD (horizontal, no slice ships value; unequal split):

```
Story 1: the whole checkout, minus error handling   (huge)
Story 2: the error messages                          (tiny)
```

GOOD (vertical walking skeleton + equal-sized follow-ups, deprioritizable):

```
Story 1 (walking skeleton): request → match → charge saved card → show "Paid"
Story 2 (Variations in Data):  multi-currency charging
Story 3 (Business Rule):       split-fare
Story 4 (Deferred NFR):        sub-200ms charge confirmation
```

Stories 2–4 are independently shippable and can be reordered or dropped by value.

## Cross-references

- SPIDR, the compact five-technique starter set this flowchart subsumes — [splitting-spidr.md](splitting-spidr.md).
- INVEST, the gate a story must fail (Small) before you split it — [invest-and-smart.md](invest-and-smart.md).
- The FDD feature list and domain walkthrough, a different decomposition lens — **fdd-guide**.
