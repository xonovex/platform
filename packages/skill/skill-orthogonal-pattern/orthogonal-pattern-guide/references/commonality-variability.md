# Shared-Core and Per-Variant Leaves

At every axis, separate what all variants share (the contract) from what each varies (one leaf). This is commonality/variability analysis. Split the system into axes first ([finding-axes.md](finding-axes.md)), then split each into core plus variants.

## Split each axis into core and variants

- **Shared = interface + types + policy** — the port variants implement, the data types they exchange, the policy/validation true for every one. Put it once at the node.
- **Each variant is a leaf** — implements the port, adds only what is unique. If two leaves need the same thing, it is commonality — sink it into `shared/`, don't copy.
- **Pick-one-per-axis is Strategy** — selecting one format/sink/compression is choosing one interchangeable family member behind one port. Callers depend on the port, never a named adapter.

## The component-cohesion principles (owned here)

- **Common Closure Principle (CCP) — a variant leaf is the closure unit.** Things that change together live together (SRP at component scope). One leaf gathers everything that changes when _that one value_ changes; a zstd shift moves only `compression/zstd/`.
- **Common Reuse Principle (CRP) — the test for `shared/`.** A type/helper enters `shared/` only if _every_ variant uses it; otherwise the non-reusers are dragged along on its every change. CRP gates the kernel; CCP gates the leaf.
- **Reuse/Release-Equivalence Principle (REP) — `shared/` is a release boundary.** Whatever sits in `shared/` is versioned and released as one thing; a change forces a re-release every leaf must absorb. That cost bounds over-factoring — factor only when reuse pays for the coupling.

```
GOOD  compression/
        shared/   # CRP gate: only what EVERY variant reuses; REP release unit
        none/ gzip/ zstd/   # each a CCP unit: all that-variant-specific change lives here
```

## Where shared code lives, kept lean

- **A per-axis `shared/`** — port, shared types, common helpers, policy. Variants sit beside it as peers.
- **Variants import `shared/`, never each other** — a leaf importing a sibling leaf is content coupling, the worst rung (**connascence-guide**). Fix: promote to `shared/`, or localize at a bridge.
- **God-package smell** — when `shared/` grows leaf-specific branches (`if s3 … else stdout …`), variability leaked upward; move the branch back down. The kernel must never name a variant.

```
BAD   sink/s3/ imports sink/file        # leaf depends on a sibling leaf
GOOD  sink/{shared,s3,file,stdout}/     # every leaf imports sink/shared only
```

## Deep, not shallow: the quality bar

A module's benefit is its functionality, its cost is its interface — make benefit outweigh cost. A variant/port should be a **deep module**: real functionality behind a narrow interface (the `Encoder` port hides parquet's schema mapping, column buffering, row-group flush behind one `Encode(record)`). A **shallow leaf** whose interface is about as complex as what it hides earns nothing — if a split yields tiny leaves that each forward a parameter, merge them or rethink the axis.

```
BAD   compression/{passthrough,identity}/   # one bool behind a full port: shallow
GOOD  compression/{shared,gzip,zstd}/        # gzip owns window/dictionary/flush; zstd owns frame+level: deep
```

## Rationale, symmetry, and fan-out

- **Give every axis explicit rationale** — record beside the port _where_ it lives, _why_ it varies, _what_ it depends on. An axis with no written reason to vary is speculative generality.
- **Treat all variants symmetrically** — same placement for every variant; do not inline two in the core file and give the third its own package (asymmetry hides a seam and invites special-casing). See [naming-symmetry.md](naming-symmetry.md).
- **Share at every level of the fan-out** — a shared-core lib feeds a shared domain lib feeds each consumer; at every hop, a shared part plus its variants. CCP/CRP/REP apply at every hop. A type two consumers reuse sinks to the shared lib so neither owns it.

```
shared-core -> shared-export (axis ports + neutral Record) -> {export-cli, export-service}  # each adds only its own adapters
```

## Don't multiply variants across axes

Producer emits neutral data; consumer applies it. An encoder emits a `Record` (bytes, content-type, schema, row count); each sink writes it its own way. Neither axis names the other, so the (sink, encoder) pairings never appear in code: **N+M leaves, not N×M**. Passing a neutral value aims at data coupling, the weakest rung (**connascence-guide**). Specifics that cannot reduce to neutral data go to a bridge ([variation-point-bridges.md](variation-point-bridges.md)).
