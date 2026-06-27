# Variation-Point Bridges

Variation points (axes) are rarely _perfectly_ independent. One variant on one
axis sometimes needs specifics produced by another. That dependency is a
**bridge** — the seam where two otherwise-independent variation points touch. The
goal is not "no coupling"; it is "no _leaked_ coupling": localize the glue at one
bridge so both axes keep varying on their own. See [finding-axes.md](finding-axes.md)
for splitting the axes in the first place, **connascence-guide**
for the connascence terms used below, and [SKILL.md](../SKILL.md) for the map.

## Why one bridge: information leakage

- **Information leakage is the why** — a single design decision (one variation point) that shows up across multiple variants is _leaked_ when each variant re-encodes it. The leaked decision is now connascent in N places; change it once and N leaves must move together. The bridge exists to confine that one decision to one site so the connascence stays local.
- **The bridge is information hiding applied to a seam** — name the leaked decision, give it one owner, and let every other leaf see only a neutral contract. "Compressors need what the encoder produced" is one decision; encode it once at the bridge, not re-derived in `gzip`, `zstd`, and `none` separately.
- **Contrast with temporal decomposition** — splitting work into "first encode, then compress" phases and then wiring those phases together is a different trap: it slices by execution order, not by reason-to-change, and re-leaks the same decision across phase boundaries. [finding-axes.md](finding-axes.md) owns that phase-splitting trap; the cure here is the same — hide the decision behind one bridge, don't spread it along a timeline.

## Spot the variation point

- **Cross-axis need is normal** — a variant on one axis often requires another axis's output. In the export machinery, `format/parquet` must lay out the column blocks a specific compressor expects, and `sink/s3` needs the content-type that the format axis produces. Neither is a flaw; each is a variation point to name and place.
- **Rule of thumb: two axes touched = a bridge** — if removing one feature would force edits on _two_ axes, the seam between them is a bridge. Name it, place it in the more-dependent leaf, and keep both axis ports ignorant of each other.

## Cross-tree constraint vs cross-cutting concern

- **Cross-tree constraint is the precise name for pairwise coupling** — when one variant _requires_ or _excludes_ a specific variant on another axis (`format/parquet` requires `compression/zstd`; `sink/stdout` excludes some buffered format), that is a cross-tree constraint between two leaves. It is bounded: exactly two variants, one edge.
- **Localize a cross-tree constraint at a bridge in the leaf** — because it touches only two variants, the glue belongs _in_ the dependent leaf (`internal/format/parquet/zstd.go`), not anywhere global. A new format that has no zstd constraint carries no such file. The constraint never reaches the variants it does not concern.
- **A cross-cutting concern is the opposite shape** — when one concern touches _every_ variant on _every_ axis (logging, the encryption gate, tracing), it is not pairwise; centralize it at the composition root, not in any leaf. [cross-cutting-concerns.md](cross-cutting-concerns.md) owns that case. The test: pairwise and bounded → bridge in the leaf; all-axis and ambient → composition root.

```
Cross-tree constraint (pairwise)   -> bridge file in the dependent leaf
  format/parquet requires compression/zstd  ->  internal/format/parquet/zstd.go

Cross-cutting concern (all-axis)   -> composition root, applied uniformly
  encryption gate / logging / tracing       ->  registry.go (see cross-cutting-concerns.md)
```

## Place the bridge in the leaf that needs it

- **Coupling flows DOWN into a leaf, never UP into a new sibling** — the bridge lives _with_ the variant that has the dependency, as a file inside that variant's directory. The format depends on a compressor, so the glue belongs in `internal/format/parquet/zstd.go` — not in a brand-new top-level `internal/export/parquetzstd/` or `zstdbridge/` package.
- **Acyclic + stable direction backs the rule** — the Acyclic Dependencies Principle forbids a cycle between the two axes, and the Stable Dependencies Principle says depend in the direction of stability: the volatile format points at the more-stable neutral contract, never the reverse. This is the connascence rule of locality made structural — keep connascent things in the same leaf and the dependency edge one-way. See **connascence-guide**.
- **The dependent variant owns the adapter** — `parquet` knows it wants a zstd block writer; `csv` knows it wants a streaming compressor. Each carries its own thin bridge file. The compression axis stays unaware of who consumes it.

```
GOOD  internal/format/{shared,json,csv,parquet}/
      internal/format/parquet/zstd.go     # bridge: parquet consumes a zstd []byte writer
      internal/format/csv/gzip.go          # bridge: csv consumes a gzip []byte writer
      internal/compression/{shared,zstd}/  # compression is its own axis, ignorant of formats

BAD   internal/export/parquetzstd/         # coupling hoisted UP into a new sibling
      internal/export/{json,csv,parquet}   # 'parquetzstd' now sits beside formats as if it were one
```

## Bridge, don't fuse

- **Keep the two abstractions decoupled** — the producing axis emits data; the consuming axis applies it. The encoder _produces_, the compressor _applies_. The bridge is the thin adapter mapping one onto the other — it does not merge them into one type. An abstraction and its implementation vary independently across the seam.
- **A fused type kills both axes' independence** — once format logic reaches into compression internals (or vice versa), you can no longer add a compressor without touching every format. Fusion converts a single bounded constraint into connascence smeared across the tree. The bridge exists precisely so each axis still has one reason to change.

## Hand off neutral data

- **Neutral-data handoff keeps the bridge thin** — the producing axis emits a neutral encoded `Record` (or `[]byte`): the serialized rows, the content-type, the column schema, the batch boundaries. It names _what to apply_, not _who applies it_. Every compressor can apply any encoded record; every encoder can target any compressor. Neutral data demotes the seam to connascence of name only.
- **Neutral data is what stops N\*M code** — with N formats and M compressors there are N*M valid combinations, but a neutral contract means N produce-sites and M apply-sites, not N*M bespoke bridges. The cross-product lives in _configuration_, not in code.

```go
// Neutral handoff: format produces this; any compressor consumes it.
type Record struct {
    Bytes       []byte   // serialized rows
    ContentType string   // MIME type for the sink
    Schema      []Column // column schema, if any
    Batches     []int    // batch boundary offsets
}

// format/parquet:    func Encode(rows Rows) (Record, error)
// compression/zstd:  func Apply(r Record, w *zstdWriter)   // in parquet/zstd.go
// compression/gzip:  func Apply(r Record, w *gzipWriter)   // in csv/gzip.go
```

```
BAD   func WriteParquetWithZstd(z *ZstdStore, blk ColumnBlock) { ... }
      // parquet reaches into compression internals; a new compressor forces a new
      // parquet function, and a new format forces a new zstd function — N*M code.

GOOD  r, _ := format.Encode(rows)   // produce once, neutrally
      zstd.Apply(r, w)              // any compressor applies the same record
```

## Anti-pattern: the tool masquerading as an axis

- **Don't promote a tool to a top-level sibling** — hoisting the s3 uploader into `internal/export/s3uploader/` makes a _tool_ look like an _axis_. `s3uploader` ends up sitting beside `json` and `csv` as though "s3-uploading" were a format, and sink logic scatters across the format tree — re-leaking the decision the bridge was meant to hide.
- **The axis is the noun; the tool is the leaf** — the sink is the axis, so it is `sink/s3` (see [naming-symmetry.md](naming-symmetry.md)); the per-format glue is the bridge file `internal/format/parquet/zstd.go`. One axis owns the capability; the dependent leaf owns the adapter. Neither earns a homeless top-level sibling.

The shared core that the bridge draws on — neutral record types, common apply
helpers — belongs to whichever axis defines it, in that axis's `shared/`; never in a
homeless `internal/exportutil` whose helpers really belong in each axis's `shared/`. See
[commonality-variability.md](commonality-variability.md) for splitting the kernel from
the variants the bridge wires together, and [structure-isomorphism.md](structure-isomorphism.md)
for keeping the bridge legible in the tree.
