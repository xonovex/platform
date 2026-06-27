# Applying the Layout

Turning variation points into a tree, or refactoring a tangled tree into variation points. The shape is always the same fan-out: a primitive core, a domain lib that owns each variation point's shared types, then each consumer realizing those points in its own medium. Decide the variation points first ([finding-axes.md](finding-axes.md)); this file places them on disk. A variation point is one axis; a variant is one leaf of it.

- **Fan out, share at each node** — `core lib → domain lib → {consumer A, consumer B}`. The core lib knows no variation point; the domain lib owns each point's shared types, policy, and source resolution; each consumer adds only its realization. Symmetry across consumers is a constraint you maintain, not luck.
- **Every node splits `shared/` from variants** — at the domain lib and inside each consumer, a variation-point dir holds a `shared/` (the commonality, [commonality-variability.md](commonality-variability.md)) plus one bare leaf per variant ([naming-symmetry.md](naming-symmetry.md)). A point in the system is one variant picked per variation point.

## The fan-out, end to end

```
shared-core-go/                      # primitives; zero variation-point knowledge
shared-export-go/                    # domain lib: per-point shared TYPES + policy + source resolution
  format/      (types, policy)       # one package per variation point, shared contract only
  sink/        (types, policy)
  compression/ (types, codec-resolve)
                 │
      ┌──────────┴──────────────────────────┐
cli consumer  (points → streaming pipeline)  service consumer  (points → scheduled job spec)
  internal/format/{shared,json,csv,parquet}      internal/builder/encoder              ~ format
  internal/sink/{shared,file,s3,stdout}          internal/builder/{writer,upload}      ~ sink
  internal/compression/{shared,none,gzip,zstd}   internal/builder/codec                ~ compression
  internal/select.go  (composition root)         internal/builder/*  (composition root)
```

Both consumers name the same three variation points and import the same shared types; only the realization (pipeline vs. job spec) differs. The chosen binding time decides how that composition root is wired — compile-time selection links one variant in, run-time selection resolves it from config through a registry (**microkernel-pattern-guide**).

## Migration: current → target

| current                          | target                                          | action   | reason                                         |
| -------------------------------- | ----------------------------------------------- | -------- | ---------------------------------------------- |
| `internal/export/s3uploader`     | `internal/sink/s3`                              | move     | sink is its own variation point, not a format  |
| `internal/exportutil`            | `internal/compression/shared` (+ `sink/shared`) | split    | each point owns its utils; no homeless sibling |
| inline `none`/`gzip` compressors | `internal/compression/{none,gzip}`              | move     | symmetric placement with the `zstd` leaf       |
| `internal/export`                | `internal/format`                               | rename   | name the variation point, not the medium       |
| `registry` / `Select` scattered  | one `select.go` composition root                | collapse | a single wiring point per consumer             |

```
BAD  internal/export/{json,csv,parquet,s3uploader}  # point named for a medium; sink misfiled inside it
     internal/exportutil/                           # utils homeless, sibling to the point instead of inside it

GOOD internal/format/{shared,json,csv,parquet}      # variation point = format; bare leaves; utils in shared/
     internal/sink/{shared,file,s3,stdout}          # sink is its own sibling point, leaves symmetric
     internal/compression/{shared,none,gzip,zstd}   # sibling point, sibling grammar
```

## Service mapping (same points, scheduled realization)

| variation point | service realization                 |
| --------------- | ----------------------------------- |
| format          | `builder/encoder`                   |
| sink            | `builder/writer` + `builder/upload` |
| compression     | `builder/codec`                     |

The service emits scheduled job specs where the CLI runs a streaming pipeline, but the point names and the domain-lib shared types are common to both. When the realizations diverge, the divergence lives in the leaf; the contract above it stays identical ([structure-isomorphism.md](structure-isomorphism.md)).

## Incremental recipe (each step compiles)

1. **Name the variation points** — agree the independent dimensions before moving a file; the names become the top dirs.
2. **One dir per point** — give each variation point a dir holding `shared/` plus a leaf per variant.
3. **Push variant-specific code into leaves** — relocate each variant's code under its own leaf; do it behind the existing port so callers don't change.
4. **Sink genuinely shared types into the domain lib** — anything both consumers (or all variants) need moves up to the shared lib, not into a per-consumer file.
5. **Localize cross-point glue at bridges** — when a leaf needs another point's specifics, put the glue inside that leaf, never in a new top-level sibling (format/parquet's compression-specific writer lives at `internal/format/parquet/zstd.go`) ([variation-point-bridges.md](variation-point-bridges.md)).
6. **Make the two consumers symmetric** — align dir names and leaf sets across consumers so a reader maps one onto the other.
7. **Collapse wiring into one composition root** — fold registry/select logic into a single selection file per consumer that picks one variant per point.

## Reversal: de-abstracting a wrong boundary

A `shared/` extracted too early couples variants by connascence they don't actually share. Reverse it — it is fine to **merge** two variation points whose boundary was wrong, or to **discard** one entirely. Sunk effort is not a reason to keep a seam that does not pay.

1. **Inline the core back into each leaf** — copy the `shared/` body into every variant that used it, then delete the `shared/` package. Each leaf now stands alone, even at the cost of duplication.
2. **Trim each copy to what that leaf uses** — drop the branches and fields a given variant never touches. The copies stop being identical; that is the point — the apparent commonality was an illusion ([commonality-variability.md](commonality-variability.md), the wrong abstraction).
3. **Re-extract only the genuine remainder** — if a true, stable commonality survives across the trimmed copies, lift just that back into `shared/`. If nothing survives, the variation point was a phantom; collapse it into its parent and move on.

```
BAD  compression/shared/Compress()    # one compressor forced over zstd + gzip + none;
     // every leaf passes flags it ignores, branches it never hits — connascence of meaning
GOOD compression/zstd/compress.go     # zstd frames a stream; gzip deflates bytes; none returns input unchanged.
     compression/gzip/compress.go     # no shared core — the "commonality" was three different jobs
     compression/none/compress.go
```

Before re-extracting, re-run the when-NOT-to-add test (rule of three, YAGNI, speculative generality, [finding-axes.md](finding-axes.md)): a boundary earns its keep only when three real variants share it.

## Risks to watch

- **False modularity** — a fan-out tree on disk proves nothing if leaves still import siblings or reach into another point's internals. Directory layout is not a boundary; ownership and an enforced import graph are. Guard the structure with architecture/import tests — fitness functions that fail the build when `format/*` imports `compression/*`, when any leaf imports a sibling leaf, or when `shared/` imports a variant ([modular-code.md](modular-code.md), [boundary-alignment.md](boundary-alignment.md)). Without them you ship a distributed monolith wearing a tidy tree.
- **Import cycles** — `shared/` must never import a variant, and variants must never import each other; all coupling flows downward into leaves or sideways through a bridge. A cycle means a "shared" type is really variant-specific — push it back down. The same import test that blocks false modularity catches this.
- **Data ownership, not file location** — each variation point owns its own data and contract; if two points read and write the same struct field, they are one point pretending to be two. Enforce single ownership per point in code, not by where the file sits.
- **Test churn** — move tests with the code into the leaf they cover; keep one table-driven test per point so adding a variant is a new table row, not a new file pattern.
- **Consumer drift** — when one consumer grows a point or type the other lacks, either push it to the domain lib or justify the asymmetry out loud; silent drift is how the two trees stop mirroring each other.

Back to [SKILL.md](../SKILL.md).
