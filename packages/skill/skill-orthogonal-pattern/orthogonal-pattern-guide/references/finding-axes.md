# Finding the Axes

Before you split a system into directories or modules, name its variation points. Get them right and the layout falls out; get them wrong and every "small" change touches three places. See [SKILL.md](../SKILL.md) for where this fits.

## A variation point is a design decision likely to change

A **variation point** (the readable alias is **axis**) is a design decision likely to change, hidden behind a port so the rest of the system never sees which choice was made — this is information hiding. A **variant** (alias: **value** or **leaf**) is one settled answer to that decision. Hide the decision, expose a stable port, and swapping variants stays local.

- **Derive variation points from the difficult, change-prone decisions** — list the choices you expect to revisit: "how is a record encoded?", "where does the output land?", "how is the stream compressed?". Each hard, likely-to-move decision is a candidate variation point.
- **Do NOT derive them from processing steps or execution phases** — slicing a system by "first open, then encode, then compress, then flush" is temporal decomposition. Phases share state and all change together when a policy shifts, so each phase module ends up touching every concern — the same failure mode as a tool masquerading as an axis. Cut along decisions that vary independently, not along the order things happen.
- **A variant leaf should have functional cohesion** — everything in a leaf exists to deliver that one variant and nothing else. Anything weaker (a leaf that also reaches across to a sibling axis) signals a missing port. See the cohesion ladder in **connascence-guide**.
- **Running illustration** — a data-export tool factors into `format {json, csv, parquet}` × `sink {file, s3, stdout}` × `compression {none, gzip, zstd}`. Each is its own change-prone decision; use it as a model, not a template.

## Vocabulary map

- **Variation point = axis** — one design decision likely to change, behind a port.
- **Variant = value = leaf** — one settled answer to that decision.
- **A product = one variant per variation point** — a concrete configuration is a tuple like `(json, s3, zstd)`, one coordinate per axis. The set of all such tuples is the cross-product / feature space. See [structure-isomorphism.md](structure-isomorphism.md) for the feature-model framing and [boundary-alignment.md](boundary-alignment.md) for choosing where the boundary falls.

## Discovering the variation points

Make each candidate state the question it answers. If two candidates answer the same question, they are one axis; if one candidate needs two answers, it is two axes.

```
format        → how is a record encoded?       json | csv | parquet
sink          → where does the output land?     file | s3 | stdout
compression   → how is the stream compressed?   none | gzip | zstd
```

- **Phrase the question, not the tool** — "where does the output land?" is the variation point; "s3" is one variant of it. Naming the dir after the question keeps the leaves bare and the port reusable.
- **Distinct questions, distinct axes** — encoding, destination, and compression move for unrelated reasons, so they are three variation points. That `parquet` can both encode and incidentally compress does not merge format and compression; the questions stay distinct.

## Two tests

**Orthogonality test — can you swap one axis's value without editing another axis's code?**

```
BAD   choosing s3 forces edits in the compression module
      (s3 leaked its multipart-buffer assumption upward)
GOOD  swap file→s3 for sink; compression/format untouched
```

If swapping leaks edits, the leak is connascence crossing a port. The aim is not zero coupling — it is no leaked coupling: localize the glue at a variation-point bridge inside the variant that needs it. See [variation-point-bridges.md](variation-point-bridges.md) and the connascence forms in **connascence-guide**.

**One-axis-or-two test — do the options always change together, or independently?**

```
BAD   enum { s3Gzip, s3None, fileGzip, fileNone }
      → sink and compression collapsed into one exploded enum
GOOD  sink {file, s3} × compression {none, gzip}
      → two axes; their product is the four states, expressed once each
```

- **Always-together → one axis** — values never chosen apart belong on the same dimension; splitting them invents a phantom axis whose halves must be kept in sync.
- **Independent → two axes** — values selected separately are separate variation points; folding them into one enum forces the cross-product to be enumerated by hand and re-edited on every new value.

## When NOT to add an axis

A port and a shared core are not free: a premature one is harder to unwind than the duplication it replaced. Test a candidate before promoting it.

- **The wrong abstraction** — a shared core extracted too early gets defended with flags, params, and per-caller conditionals as each new caller bends it. Unwinding that is harder than having left the code duplicated. Symptom to detect: **if `shared/` grows flags and per-variant branches, it was never one variation point** — inline it back and let the variants diverge.
- **Rule of three** — tolerate duplication until the third occurrence. Two similar `compression` paths may be coincidence; the third real variant reveals the actual axis of change. Extract the port then, not on the first guess.
- **YAGNI** — keep the system easy to extend (a clean port the moment a second variant lands); defer code already extended for a variant nobody is building. "Easy to add `zstd` later" is cheap; a `zstd` leaf with no caller is dead weight.
- **Speculative generality** — the named smell for abstract bases, hooks, or extension points added for a presumed future variant. An `Encoder` port with one implementation and three unused override points is speculative generality, not a variation point.

Cross-check against the deep-vs-shallow quality bar in [commonality-variability.md](commonality-variability.md): a shallow port whose interface is as wide as its single implementation hides nothing and earns no axis.

## A concern at every point is not a variation point

If something is present at every variant rather than chosen between variants — logging on every encoder, an audit hook on every sink — it is not an axis. It is a cross-cutting concern, and folding it into one variant's leaf scatters it. Factor it separately; see [cross-cutting-concerns.md](cross-cutting-concerns.md).
