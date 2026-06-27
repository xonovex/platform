# Shared-Core and Per-Variant Leaves

At every variation point, separate what all variants share from what each one varies. The shared part is the contract; each variant is one leaf that fulfils it. This is commonality/variability analysis: name the common kernel once, push every difference to a variant. Split the system into variation points first — see [finding-axes.md](finding-axes.md) — then split each one into core plus variants.

A variation point (axis) is a design decision likely to change; a variant (leaf) is one value of it. The component-cohesion principles below name _why_ the split lands where it does.

## Split each variation point into core and variants

- **Shared = interface + types + policy** — the commonality is everything all variants need to stay interchangeable: the port they implement, the data types they exchange, and the policy/validation that holds for every one. Put it once at the node.
- **Each variant is a leaf** — a leaf implements the port and adds only what is unique to it. If two leaves need the same thing, it is commonality, not variation — sink it into the shared part, don't copy it.
- **Pick-one-per-variation-point is Strategy** — selecting one format, one sink, one compression is choosing one interchangeable family member behind one port. The interface is the port; each concrete type is an adapter. Callers depend on the port, never on a named adapter.

## Name the principles: CCP, CRP, REP

The heuristics above instantiate the three component-cohesion principles. They are owned here.

- **Common Closure Principle — a variant leaf is the closure unit.** Things that change together live together; this is the single-responsibility principle at component scope. One variant leaf gathers everything that changes when _that one value_ of the variation point changes — its adapter, its private types, its helpers. When a zstd detail shifts, only `compression/zstd/` should move.
- **Common Reuse Principle — the precise test for `shared/`.** Do not make variants depend on code they do not reuse. A type or helper enters `shared/` only if _every_ variant actually uses it; otherwise the leaves that don't reuse it are dragged along on its every change. CRP is the gate on the kernel; CCP is the gate on the leaf.
- **Reuse/Release-Equivalence Principle — `shared/` is a release boundary.** The unit of reuse is the unit of release: whatever sits in `shared/` is versioned and released as one thing, and every variant depends on that release. That carries a real cost — a change to `shared/` forces a re-release that every leaf must absorb — and that cost is what bounds over-factoring. Factor into `shared/` only when the reuse pays for the coupling.

```
GOOD  compression/
        shared/   # CRP gate: only what EVERY variant reuses; REP release unit
        none/     # CCP unit: all "none"-specific change lives here
        gzip/     # CCP unit
        zstd/     # CCP unit
```

## Where shared code lives

- **A variation-point-local `shared/` package** — the port, the shared data types, the common helpers, and the policy live in a `shared/` (or `core`) sibling inside the variation-point directory. Variants sit beside it as peers.
- **Variants import `shared/`; never each other** — dependency flows one way, leaf to shared. A leaf importing a sibling leaf is content coupling — the strongest, worst rung of the coupling ladder; see **connascence-guide**. The fix: promote the shared concept to `shared/`, or localize it at a bridge.

```
BAD   sink/
        s3/      imports sink/file  # leaf depends on a sibling leaf
        file/
        stdout/
GOOD  sink/
        shared/  # port, Record type, policy, helpers
        s3/      imports sink/shared
        file/    imports sink/shared
        stdout/  imports sink/shared
```

## Keep `shared/` lean

- **It holds only what every variant reuses** — `shared/` is a kernel, not a junk drawer. This is CRP again: a type or helper that exactly one leaf uses belongs in that leaf, even when it feels shared-ish.
- **God-package smell** — when `shared/` grows leaf-specific branches (`if s3 … else stdout …`), the variability leaked upward. Move that branch back down into the leaf it describes; the kernel must never name a variant.

## Deep, not shallow: the quality bar

A module's benefit is its functionality; its cost is its interface. Make the benefit outweigh the cost.

- **A variant/port should be a deep module** — pack real functionality behind a narrow interface. The Encoder port hides parquet's whole schema mapping, column buffering, and row-group flush logic behind one `Encode(record)`; that depth is what makes the cut worth it.
- **A shallow leaf is a smell** — a leaf whose interface is about as complex as what it hides earns nothing for the seam it adds. If splitting one variation point yields several tiny leaves that each just forward a parameter, the cut is wrong: merge them, or rethink the variation point. This is the quantitative guard against multiplying shallow variants — pair it with the when-NOT-to-add test in [finding-axes.md](finding-axes.md).

```
BAD   compression/
        passthrough/ # one bool behind a full port — interface ~ contents: shallow
        identity/
GOOD  compression/
        shared/      # Compressor port + []byte
        gzip/        # owns window sizing, dictionary, flush assembly: deep
        zstd/        # owns frame format + per-level tuning: deep
```

## Give every variation point explicit rationale

- **Each variation point carries its own justification** — record, beside the port, _where_ it lives, _why_ it varies (the decision likely to change), and _what it depends on_. A variation point without a written reason to vary is speculative generality waiting to be deleted. This rationale is what lets the next reader tell a load-bearing seam from an accidental one.

## Treat all variants symmetrically

- **Same placement for every variant** — do not inline two variants in the core file and give the third its own package. Asymmetric placement hides one variant's seam, invites special-casing, and breaks the parallel naming the tree relies on — see [naming-symmetry.md](naming-symmetry.md) and [structure-isomorphism.md](structure-isomorphism.md).

```
BAD   format/
        format.go   # json + csv implemented inline here
        parquet/    # only parquet gets its own package — asymmetry smell
GOOD  format/
        shared/     # port + Record
        json/
        csv/
        parquet/    # every variant is a peer leaf, placed identically
```

## Share at every level of the fan-out

- **Each level repeats the pattern** — a shared-core lib feeds a shared domain lib feeds each consumer; at every hop there is a shared part plus its variants, and a consumer adds only its own realization. CCP/CRP/REP apply at every hop, not just the leaves.
- **Genuinely cross-consumer types sink to the shared lib** — a type two consumers both reuse is commonality at that level (CRP satisfied). Move it down so neither consumer owns it and they stay symmetric.

```
shared-core            # primitives every consumer needs
  -> shared-export     # variation-point ports + the neutral Record type
    -> export-cli      # its own adapters only
    -> export-service  # its own adapters only
```

## Don't multiply variants across variation points

- **Producer emits neutral data; consumer applies it** — when one variation point produces something a second consumes, have the producer return a plain data value and let the consumer apply it through its own mechanism. An encoder emits a Record (encoded bytes, content-type, schema metadata, row count); each sink writes it its own way. Passing a neutral value rather than a shared object aims at data coupling, the weakest rung — see **connascence-guide**.
- **N+M, not N×M** — neither variation point names the other, so the (sink, encoder) pairings never appear in code: N sinks and M encoders stay N+M leaves, not N×M. Cross-axis specifics that cannot reduce to neutral data go to a bridge — see [variation-point-bridges.md](variation-point-bridges.md).

Back to the overview: [SKILL.md](../SKILL.md).
