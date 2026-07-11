# Applying the Layout

Turning axes into a tree, or refactoring a tangled tree into axes. The shape is always the same fan-out: a primitive core, a domain lib that owns each axis's shared types, then each consumer realizing those axes in its own medium. Decide the axes first ([finding-axes.md](finding-axes.md)); this file places them on disk.

- **Fan out, share at each node** — `core lib → domain lib → {consumer A, consumer B}`. The core lib knows no axis; the domain lib owns each axis's shared types, policy, and source resolution; each consumer adds only its realization. Symmetry across consumers is a constraint you maintain, not luck.
- **Every node splits `shared/` from variants** — a `shared/` (commonality, [commonality-variability.md](commonality-variability.md)) plus one bare leaf per variant ([naming-symmetry.md](naming-symmetry.md)).

```
shared-core-go/                      # primitives; zero axis knowledge
shared-export-go/                    # domain lib: per-axis shared TYPES + policy + source resolution
  format/ sink/ compression/         # one package per axis, shared contract only
      ┌──────────┴──────────────────────────┐
cli consumer  (axes → streaming pipeline)   service consumer  (axes → scheduled job spec)
  internal/format/{shared,json,csv,parquet}      internal/builder/encoder              ~ format
  internal/sink/{shared,file,s3,stdout}          internal/builder/{writer,upload}      ~ sink
  internal/compression/{shared,none,gzip,zstd}   internal/builder/codec                ~ compression
  internal/select.go  (composition root)         internal/builder/*  (composition root)
```

Both consumers name the same three axes and import the same shared types; only the realization differs. Binding time decides how the composition root is wired — compile-time selection links one variant in, run-time selection resolves it from config through a registry (**microkernel-pattern-guide**).

## Migration: current → target

| current                          | target                                          | action   | reason                                        |
| -------------------------------- | ----------------------------------------------- | -------- | --------------------------------------------- |
| `internal/export/s3uploader`     | `internal/sink/s3`                              | move     | sink is its own axis, not a format            |
| `internal/exportutil`            | `internal/compression/shared` (+ `sink/shared`) | split    | each axis owns its utils; no homeless sibling |
| inline `none`/`gzip` compressors | `internal/compression/{none,gzip}`              | move     | symmetric placement with the `zstd` leaf      |
| `internal/export`                | `internal/format`                               | rename   | name the axis, not the medium                 |
| `registry` / `Select` scattered  | one `select.go` composition root                | collapse | a single wiring point per consumer            |

## Incremental recipe (each step compiles)

1. **Name the axes** — agree the independent dimensions; the names become the top dirs.
2. **One dir per axis** — `shared/` plus a leaf per variant.
3. **Push variant-specific code into leaves** — behind the existing port so callers don't change.
4. **Sink genuinely shared types into the domain lib** — anything all variants (or both consumers) need moves up, not into a per-consumer file.
5. **Localize cross-axis glue at bridges** — inside the dependent leaf (`internal/format/parquet/zstd.go`), never a new top-level sibling ([variation-point-bridges.md](variation-point-bridges.md)).
6. **Make the two consumers symmetric** — align dir names and leaf sets so a reader maps one onto the other.
7. **Collapse wiring into one composition root** — one selection file per consumer picking one variant per axis.

## Reversal: de-abstracting a wrong boundary

A `shared/` extracted too early couples variants by connascence they don't share. It is fine to **merge** two axes whose boundary was wrong, or **discard** one; sunk effort is no reason to keep a seam that doesn't pay.

1. **Inline the core back into each leaf** — copy the `shared/` body into every variant that used it, delete the `shared/` package.
2. **Trim each copy to what that leaf uses** — the copies stop being identical; that is the point ([commonality-variability.md](commonality-variability.md), the wrong abstraction).
3. **Re-extract only the genuine remainder** — if a true, stable commonality survives, lift just that back; if nothing survives, the axis was a phantom.

```
BAD  compression/shared/Compress()   # one compressor forced over zstd+gzip+none; every leaf passes flags it ignores
GOOD compression/{zstd,gzip,none}/compress.go   # three different jobs — no shared core; the "commonality" was an illusion
```

Before re-extracting, re-run the when-NOT-to-add test (rule of three, YAGNI, speculative generality; [finding-axes.md](finding-axes.md)).

## Risks to watch

- **False modularity** — a fan-out tree proves nothing if leaves import siblings or reach into another axis's internals. Guard with architecture/import tests that fail the build when `format/*` imports `compression/*`, a leaf imports a sibling, or `shared/` imports a variant ([modular-code.md](modular-code.md), [boundary-alignment.md](boundary-alignment.md)).
- **Import cycles** — `shared/` must never import a variant, variants never each other; coupling flows down into leaves or sideways through a bridge. A cycle means a "shared" type is really variant-specific.
- **Data ownership, not file location** — each axis owns its data and contract; if two axes read and write the same struct field, they are one axis pretending to be two.
- **Test churn** — move tests into the leaf they cover; keep one table-driven test per axis so adding a variant is a new row, not a new file.
- **Consumer drift** — a point one consumer grows and the other lacks either pushes to the domain lib or is justified out loud; silent drift stops the two trees mirroring.
