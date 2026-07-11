# Variation-Point Bridges

Axes are rarely _perfectly_ independent: one variant sometimes needs specifics produced by another. That dependency is a **bridge** — the seam where two axes touch. The goal is not "no coupling" but "no _leaked_ coupling": localize the glue at one bridge so both axes keep varying on their own. See [finding-axes.md](finding-axes.md) and **connascence-guide**.

- **Information leakage is the why** — one design decision showing up across multiple variants is leaked when each re-encodes it, becoming connascent in N places. The bridge confines that decision to one site. "Compressors need what the encoder produced" is one decision — encode it once, not re-derived in `gzip`, `zstd`, and `none` separately.
- **Two axes touched = a bridge** — if removing one feature would force edits on _two_ axes, the seam between them is a bridge. Name it, place it in the more-dependent leaf, keep both axis ports ignorant of each other.

## Cross-tree constraint vs cross-cutting concern

- **Cross-tree constraint** — one variant _requires_/_excludes_ a specific variant on another axis (`format/parquet` requires `compression/zstd`; `sink/stdout` excludes some buffered format). Bounded: exactly two variants, one edge. Localize it as a bridge file _in_ the dependent leaf; a new format with no zstd constraint carries no such file.
- **Cross-cutting concern** — one concern touches _every_ variant on _every_ axis (logging, encryption gate, tracing). Centralize at the composition root, never in a leaf. See [cross-cutting-concerns.md](cross-cutting-concerns.md).

```
Cross-tree constraint (pairwise)  -> bridge file in the dependent leaf
  format/parquet requires compression/zstd  ->  internal/format/parquet/zstd.go
Cross-cutting concern (all-axis)  -> composition root, applied uniformly
  encryption gate / logging / tracing       ->  registry.go
```

## Place the bridge in the leaf that needs it

Coupling flows DOWN into a leaf, never UP into a new sibling. The format depends on a compressor, so the glue is `internal/format/parquet/zstd.go` — not a new top-level `internal/export/parquetzstd/`. This backs the Acyclic Dependencies Principle (no cycle between axes) and the Stable Dependencies Principle (the volatile format points at the more-stable neutral contract, never the reverse).

```
GOOD  internal/format/parquet/zstd.go     # bridge: parquet consumes a zstd []byte writer
      internal/format/csv/gzip.go         # bridge: csv consumes a gzip []byte writer
      internal/compression/{shared,zstd}/ # compression is its own axis, ignorant of formats
BAD   internal/export/parquetzstd/        # coupling hoisted UP into a homeless sibling
```

## Bridge, don't fuse — hand off neutral data

The producing axis emits data; the consuming axis applies it. The bridge is a thin adapter, not a merged type: once format logic reaches into compression internals you can no longer add a compressor without touching every format. The producer emits a neutral `Record` (serialized rows, content-type, schema, batch boundaries) naming _what to apply_, not _who applies it_ — demoting the seam to connascence of name. With N formats and M compressors, a neutral contract means N produce-sites and M apply-sites, not N*M bespoke bridges; the cross-product lives in _configuration_, not code.

```go
type Record struct {
    Bytes       []byte   // serialized rows
    ContentType string   // MIME type for the sink
    Schema      []Column // column schema, if any
    Batches     []int    // batch boundary offsets
}
// format/parquet:   func Encode(rows Rows) (Record, error)
// compression/zstd: func Apply(r Record, w *zstdWriter)   // lives in parquet/zstd.go
```

```
BAD   func WriteParquetWithZstd(z *ZstdStore, blk ColumnBlock) { ... }  # N*M code
GOOD  r, _ := format.Encode(rows); zstd.Apply(r, w)                     # any compressor applies the same record
```

## Anti-pattern: the tool masquerading as an axis

Hoisting the s3 uploader into `internal/export/s3uploader/` makes a _tool_ look like an _axis_, sitting beside `json`/`csv` as though "s3-uploading" were a format. The axis is the noun (`sink/s3`, see [naming-symmetry.md](naming-symmetry.md)); the per-format glue is the bridge file. Shared neutral types belong to whichever axis defines them, in that axis's `shared/`, never a homeless `internal/exportutil`. See [commonality-variability.md](commonality-variability.md).
