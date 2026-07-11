# Structure Mirrors the Model

Make the directory tree isomorphic to the conceptual model so reading it reveals the axes and variants with no external map — the structure should scream the domain, not the layer. See [SKILL.md](../SKILL.md).

## Package by variation point, not by layer

- **Top dir per axis** — name each directory for one design decision likely to change. A newcomer reads the dirs and recovers the model.
- **Leaf per variant, plus `shared/`** — each axis dir holds bare variant siblings plus a `shared/` core. See [naming-symmetry.md](naming-symmetry.md).
- **Each axis owns its own types/utils/shared** — a root full of `utils/`, `helpers/`, `types/`, `services/`, `controllers/` smears one axis across many dirs (package-by-layer) and hides the axes. Push every helper down into the axis that uses it.

```
BAD  package-by-layer — screams the framework, hides the model
internal/{types,utils,services,controllers}/
GOOD package-by-axis — screams the domain
internal/
  format/{shared,json,csv,parquet}/    # how a record is encoded
  sink/{shared,file,s3,stdout}/        # where the bytes land
  compression/{shared,none,gzip,zstd}/ # how the bytes are squeezed
  registry.go                          # single composition root
```

## Ports and adapters

- **Each axis interface is a port** — the abstract contract lives in that axis's `shared/`; consumers depend on the port, never a concrete leaf (the microkernel structure).
- **Each concrete variant is an adapter** — `sink/s3/` adapts object storage to the `Sink` port. Swapping variants is Strategy; bridging a port to a foreign API is the adapter side of hexagonal.
- **Consumers are adapters too** — a CLI and a scheduled service are two driving adapters over the same ports; anything both need lives below them in `shared/`. See [applying-the-layout.md](applying-the-layout.md).

## One composition root

A single file (a `DefaultRegistry`) is the only thing that imports every concrete variant and binds it to its port; everything else depends on interfaces only. That one-way fan-in keeps the graph acyclic and the swap surface tiny.

```go
// registry.go — the ONLY importer of concrete variants
func DefaultRegistry() Registry {
  return Registry{
    Format:      map[string]Encoder{"json": json.New(), "csv": csv.New(), "parquet": parquet.New()},
    Sink:        map[string]Sink{"file": file.New(), "s3": s3.New(), "stdout": stdout.New()},
    Compression: map[string]Compressor{"none": none.New(), "gzip": gzip.New(), "zstd": zstd.New()},
  }
}
```

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
