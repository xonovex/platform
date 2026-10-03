# user-stories: Splitting Stories with SPIDR

SPIDR is the compact five-technique starter set for splitting a too-big story into smaller stories that each still deliver value. Reach for the fuller decision tree in [splitting-flowchart.md](splitting-flowchart.md) when SPIDR doesn't yield a clean split.

- **S: Spike**: research-only, produces knowledge not shippable value; a spike is **not** itself a vertically-sliced story and is the **last resort**. Resolve the uncertainty, then split with another technique.
- **P: Path**: split by the paths a user can take; ship one path end to end first.
- **I: Interface**: ship the plainest usable interface first (singular "Interface"), enrich later.
- **D: Data**: support a data subset first (one currency/region/format), add the rest later.
- **R: Rules**: relax a business rule for the first slice, reintroduce deferred rules as their own stories.

Each technique carves a thin vertical slice, not a horizontal layer. Every resulting story must still pass INVEST, especially **Valuable** and **Small** (see [invest-and-smart.md](invest-and-smart.md)).

## Example (parent: "pay for my completed trip")

```
Story 1 (Rules + Data): auto-charge the saved card, one currency  ← shippable value
Story 2 (Path):         add the wallet payment path
Story 3 (Rules):        add split-fare
Story 4 (Interface):    add the animated itemized receipt
```

Contrast the horizontal non-split ("build the API / DB / UI"): none delivers value alone.

## Cross-references

- The full pattern flowchart and vertical-slicing / walking-skeleton concept: [splitting-flowchart.md](splitting-flowchart.md).
- The INVEST properties every split story must still pass: [invest-and-smart.md](invest-and-smart.md).
