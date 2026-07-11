# Finding the Axes

Name the variation points before splitting into directories. A **variation point** (alias **axis**) is a design decision likely to change, hidden behind a port (information hiding); a **variant** (alias **value**, **leaf**) is one settled answer. A **product** is one variant per axis — a tuple like `(json, s3, zstd)`, one coordinate per axis; the set of all tuples is the cross-product / feature space. See [SKILL.md](../SKILL.md).

- **Derive axes from change-prone decisions, never from processing steps** — list the choices you expect to revisit ("how is a record encoded?", "where does the output land?", "how is the stream compressed?"). Slicing by execution order ("open, then encode, then compress, then flush") is temporal decomposition: phases share state and all change together, so each phase module touches every concern.
- **A variant leaf has functional cohesion** — everything in it exists to deliver that one variant; a leaf reaching across to a sibling axis signals a missing port (cohesion ladder in **connascence-guide**).
- **Running illustration** — a data-export tool factors into `format {json, csv, parquet}` × `sink {file, s3, stdout}` × `compression {none, gzip, zstd}`. Model, not template.

## Discovering axes

Make each candidate state the question it answers. Same question → one axis; one candidate needing two answers → two axes. Name the dir for the question, not the tool (`sink`, not `s3`). That `parquet` incidentally compresses does not merge `format` and `compression`; the questions stay distinct.

## Two tests

**Orthogonality — can you swap one axis's value without editing another axis's code?**

```
BAD   choosing s3 forces edits in the compression module (s3 leaked its multipart-buffer assumption upward)
GOOD  swap file→s3 for sink; compression/format untouched
```

A leak is connascence crossing a port; the aim is not zero coupling but no _leaked_ coupling — localize the glue at a variation-point bridge in the variant that needs it. See [variation-point-bridges.md](variation-point-bridges.md).

**One-axis-or-two — do the options always change together, or independently?**

```
BAD   enum { s3Gzip, s3None, fileGzip, fileNone }   # sink and compression collapsed into one exploded enum
GOOD  sink {file, s3} × compression {none, gzip}    # two axes; their product is the four states, once each
```

Always-together → one axis (splitting invents a phantom axis whose halves must stay in sync). Independent → two axes (folding into one enum forces the cross-product to be hand-enumerated and re-edited on every new value).

## When NOT to add an axis

A port and shared core are not free; a premature one is harder to unwind than the duplication it replaced.

- **The wrong abstraction** — if `shared/` grows flags and per-variant branches, it was never one axis: inline it back and let the variants diverge.
- **Rule of three** — tolerate duplication until the third real variant reveals the actual axis; extract the port then.
- **YAGNI** — a clean port the moment a second variant lands is cheap; a `zstd` leaf with no caller is dead weight.
- **Speculative generality** — an `Encoder` port with one implementation and three unused override points is not a variation point.

Cross-check against the deep-vs-shallow bar in [commonality-variability.md](commonality-variability.md): a shallow port as wide as its single implementation hides nothing and earns no axis.

A concern present at _every_ variant (logging on every encoder, an audit hook on every sink) is not an axis — it is a cross-cutting concern; folding it into one leaf scatters it. See [cross-cutting-concerns.md](cross-cutting-concerns.md).
