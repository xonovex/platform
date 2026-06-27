# Structure Mirrors the Model

The directory tree is the diagram. Make it isomorphic to the conceptual model so that reading the tree reveals the variation points and their variants with no external map. This is the screaming-architecture test: the structure should scream the domain, not the framework or the layer it sits in.

## Package by variation point, not by layer

- **Top dir per variation point** — name each directory for a variation point (= axis): one design decision likely to change, with one reason to change. A newcomer reads the dirs and recovers the model. See [finding-axes.md](finding-axes.md) for picking them.
- **Leaf per variant** — each axis dir holds its concrete variants (= leaves) as bare siblings, plus a `shared/` for the common core. See [naming-symmetry.md](naming-symmetry.md) for keeping leaf names parallel.
- **Each axis owns its own types/utils/shared** — a root full of `utils/`, `helpers/`, `types/`, `controllers/`, `services/` smears one axis across many dirs (package-by-layer) and hides the variation points entirely. Push every helper down into the axis that uses it.

```
BAD  package-by-layer — the tree screams the framework, hides the model
internal/
  types/        # one format type here, one sink type there...
  utils/        # a csv helper next to an s3 helper next to a gzip helper
  services/     # all the "do the thing" verbs, axis-blind
  controllers/  # cli + scheduler glue tangled together
```

```
GOOD package-by-axis — the tree screams the domain; variation points and variants are legible
internal/
  format/{shared,json,csv,parquet}/    # variation point: how a record is encoded
  sink/{shared,file,s3,stdout}/        # variation point: where the bytes land
  compression/{shared,none,gzip,zstd}/ # variation point: how the bytes are squeezed
  export/{shared,...}/                 # variation point: how the run is driven
  registry.go                          # the single composition root (below)
```

## Ports and adapters

- **Each axis interface is a port** — the abstract contract for "an encoder", "a sink" lives in that axis's `shared/`; consumers depend on the port, never on a concrete leaf. This is the microkernel structure: a minimal core plus variants behind a port.
- **Each concrete variant is an adapter** — `sink/s3/` adapts object storage to the `Sink` port; `format/parquet/` adapts columnar encoding to the `Encoder` port. Swapping variants is Strategy; bridging a port to a foreign API is the adapter side of hexagonal architecture.
- **Consumers are adapters too** — a CLI and a scheduled service are just two driving adapters over the same shared ports. Keep them symmetric: anything both need lives below them in `shared/`, not duplicated in each. See [applying-the-layout.md](applying-the-layout.md).

```
shared-core
  └─ shared-lib            # ports + shared/ cores for every axis
       ├─ cli (adapter)    # selects variants, wires the composition root
       └─ scheduler (adapter)  # same ports, same registry shape
```

## One composition root

- **Exactly one wiring place** — a single file (e.g. a `DefaultRegistry`) is the only thing that imports every concrete variant and binds it to its port. This is the composition root.
- **Everything else depends on interfaces only** — leaves know their port and nothing about siblings; the registry knows all leaves but no business logic. That one-way fan-in keeps the dependency graph acyclic and the swap surface tiny.

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

Make the placement test measurable. Instability is `I = Ce / (Ca + Ce)` — outgoing dependencies over total. A node nothing depends on but that depends on much is I≈1 (free to change); a node everything depends on but that depends on nothing is I≈0 (rigid, must stay still).

- **The port/core is the most-depended-on node** — it has high `Ca`, so drive it to LOW instability and keep it ABSTRACT. A stable thing may only be safely stable if it is abstract: the Stable Abstractions Principle. The `Encoder` port and the neutral encoded `Record` value are depended on by every leaf and every consumer, so they must be interfaces and value types, not concretions.
- **Variant leaves are the most-changeable nodes** — `s3/`, `parquet/`, `zstd/` are depended on by nothing but the registry, so they sit at high instability and are free to be concrete and churn.

```
            Ca (depended-on)        Ce (depends-on)      I = Ce/(Ca+Ce)
Encoder port       high                  ~0                 ~0  abstract, stable  ✓
s3 leaf            ~0 (registry only)    high               ~1  concrete, volatile ✓
exportutil/        high                  high               ~0.5 stable+concrete  ✗ Zone of Pain
```

- **A stable-yet-concrete shared module is the Zone of Pain** — depended on by many (rigid) but concrete (begs to change). An `exportutil/` that every axis imports yet hard-codes s3 assumptions cannot be edited without breaking its dependents. Fix it by extracting an abstract port for the part that varies and demoting the concrete tail into a single leaf.
- **The grep test** — if the most-depended-on file in an axis is concrete (no interface, imports a specific tool), it is drifting into the Zone of Pain. Keep the fan-in target abstract.

## The cross-product is a feature model

The valid configurations are not the full cartesian product of the leaves; they are a constrained subset — a feature model. The tree shows the axes and variants; the feature model adds which combinations are legal.

- **Variants narrowed by cross-tree constraints** — `requires`/`excludes` rules cut the product down (e.g. `parquet` format requires a `Compressor` it can frame internally, `stdout` sink excludes the multipart upload). These constraints span axes and live at a single variation-point bridge, not smeared across leaves. See [variation-point-bridges.md](variation-point-bridges.md).
- **Variability is its own dimension (orthogonal variability model)** — document the variation as a layer cross-cutting the base model, not buried inside the layers. One top-level dir per axis IS that model rendered as a tree: the variability is legible as its own dimension instead of hidden in a `services/` pile.

```
feature model (legal points), not the raw 3×3×3 product:
  format=parquet   requires  compression∈{none,zstd}    # parquet frames internally
  sink=stdout      excludes  upload=multipart           # no object store, no multipart
```

A fail-closed policy gate enforces the feature model at the composition root: an illegal tuple is rejected, not silently run. See **microkernel-pattern-guide** for the gate; [finding-axes.md](finding-axes.md) for reading a config as coordinates.

## Some concerns live at no node

Not every concern maps to a directory. A cross-cutting concern — logging, the policy gate, audit — appears at NO single node of the tree because it threads through many. Its absence from the layout is intentional, not a modelling gap: it is woven across the axes rather than owned by one. Do not invent a `logging/` axis to make it visible; see [cross-cutting-concerns.md](cross-cutting-concerns.md) for why it stays cross-cutting and how it is woven.

## One owner per variation point

- **The team shape and the tree shape converge (Conway's law)** — assign each axis a single owner so its `shared/` port and its leaves move together under one hand; split ownership and the port drifts from its adapters. Depth and the enforcement tests live in [boundary-alignment.md](boundary-alignment.md).

## The tree as documentation

- **Inferable extension** — a newcomer should deduce "to add a parquet encoder, add `format/parquet/` and register it once" purely from the shape, without reading a contributing guide.
- **Isomorphism test** — if they cannot infer where a new variant goes, or which axis owns a concern, the structure is not isomorphic to the model. Fix the tree, not the docs.
- **Smell** — a dir named for a tool or a layer instead of an axis (`s3uploader/`, `exportutil/`, `services/`) means the model leaked. The axis is the noun; the tool is the leaf; the layer is nobody.

For the coupling and cohesion vocabulary used throughout — connascence, the coupling and cohesion ladders, the Law of Demeter — see **connascence-guide**. See [SKILL.md](../SKILL.md) for how this fits the other axes of decomposition.
