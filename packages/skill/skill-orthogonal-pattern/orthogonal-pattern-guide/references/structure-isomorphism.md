# Structure Mirrors the Model

Make the directory tree isomorphic to the conceptual model so reading it reveals the axes and variants with no external map — the structure should scream the domain, not the layer. Name each top dir for one axis, hold bare variant siblings plus a `shared/` core, and let each axis own its own types/utils rather than a root-level `utils/`/`helpers/`/`services/`. See [SKILL.md](../SKILL.md) and [naming-symmetry.md](naming-symmetry.md).

## Ports and adapters

- **Each axis interface is a port** — the abstract contract lives in that axis's `shared/`; consumers depend on the port, never a concrete leaf (the microkernel structure).
- **Each concrete variant is an adapter** — `sink/s3/` adapts object storage to the `Sink` port. Swapping variants is Strategy; bridging a port to a foreign API is the adapter side of hexagonal.
- **Consumers are adapters too** — a CLI and a scheduled service are two driving adapters over the same ports; anything both need lives below them in `shared/`. See [applying-the-layout.md](applying-the-layout.md).

## Stable abstract core, unstable concrete leaves

Instability `I = Ce / (Ca + Ce)` — outgoing deps over total. The port/core has high `Ca`, so drive it to LOW instability and keep it ABSTRACT (Stable Abstractions Principle: a stable thing is only safely stable if abstract). Variant leaves are depended on by nothing but the registry, so they sit at high I and are free to be concrete and churn.

```
            Ca (depended-on)     Ce (depends-on)   I = Ce/(Ca+Ce)
Encoder port    high                 ~0               ~0   abstract, stable   ✓
s3 leaf         ~0 (registry only)   high             ~1   concrete, volatile ✓
exportutil/     high                 high             ~0.5 Zone of Pain       ✗
```

A **stable-yet-concrete** shared module is the **Zone of Pain** — depended on by many (rigid) but concrete (begs to change). Fix by extracting an abstract port for the varying part and demoting the concrete tail into a single leaf. Grep test: if the most-depended-on file in an axis is concrete (no interface, imports a specific tool), it is drifting into the Zone of Pain.

## The cross-product is a feature model

Valid configurations are a constrained subset (a feature model), not the full cartesian product. `requires`/`excludes` rules span axes and live at a single variation-point bridge, not smeared across leaves. See [variation-point-bridges.md](variation-point-bridges.md).

```
feature model (legal points), not the raw 3×3×3 product:
  format=parquet   requires  compression∈{none,zstd}    # parquet frames internally
  sink=stdout      excludes  upload=multipart           # no object store, no multipart
```

A fail-closed policy gate enforces it at the composition root: an illegal tuple is rejected, not silently run (**microkernel-pattern-guide**).

## Concerns at no node, one owner, tree as docs

- **Some concerns live at no node** — a cross-cutting concern (logging, policy gate, audit) appears at NO directory because it threads through many. Its absence is intentional; do not invent a `logging/` axis. See [cross-cutting-concerns.md](cross-cutting-concerns.md).
- **One owner per axis (Conway's law)** — assign each axis a single owner so its `shared/` port and leaves move together; depth and enforcement tests in [boundary-alignment.md](boundary-alignment.md).
- **Isomorphism test** — a newcomer should deduce "add `format/parquet/` and register it once" purely from the shape. If they cannot infer where a new variant goes, or a dir is named for a tool/layer (`s3uploader/`, `exportutil/`, `services/`), the model leaked — fix the tree, not the docs.

For connascence, coupling/cohesion ladders, and the Law of Demeter see **connascence-guide**.
