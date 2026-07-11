# Naming and Symmetry

Parallel concepts deserve parallel names: when siblings are named consistently, a reader who learns one guesses the rest. A directory names an _axis_ (a noun — the question); a leaf names a _variant_ (one value). See [structure-isomorphism.md](structure-isomorphism.md) for the tree, [commonality-variability.md](commonality-variability.md) for what lives in each node.

## Sibling concepts get sibling names

Siblings under one parent answer the same kind of question at the same altitude. An axis is a noun (`format`, `sink`, `compression`); a dir named for a verb or mechanism (`encode`, `upload`) reads as a different category and breaks the parallel. One odd sibling poisons the set — rename `encode/` to the axis it realizes (`format`).

```
BAD  internal/{encode, sink, compression}/   # 'encode' is a verb, the rest are axes
GOOD internal/{format, sink, compression}/   # three nouns, three questions, one altitude
```

## The axis dir supplies context — keep the leaf bare

`sink/s3` reads as "the s3 variant of sink"; suffixing to `sink/s3uploader` repeats the parent. Name the leaf for the variant, nothing more. A homeless top-level `exportutil/` means an axis's own helpers leaked out — push them into `sink/shared/`.

| BAD (leaf repeats axis)       | GOOD (bare leaf)   |
| ----------------------------- | ------------------ |
| `sink/s3uploader`             | `sink/s3`          |
| `format/jsonformat`           | `format/json`      |
| `sink/stdoutsink`             | `sink/stdout`      |
| `compression/gzipcompression` | `compression/gzip` |
| `sink/filewriter`             | `sink/file`        |

## Options are namespaced by axis + variant

A per-variant option carries its axis and variant as a prefix, so the flag namespace is isomorphic to the tree. Namespaced named options weaken connascence: positional flags impose connascence of position (caller and parser must agree on order; a reorder silently breaks them), a name imposes connascence of name — the weaker, more local form (rule of degree). The prefix also makes each name unique, so `--level` for two variants never collides. Adding a variant adds a leaf plus its own `--<axis>-<variant>-*` options, not an edit to a god-parser. See **connascence-guide**.

```
BAD   --bucket ...  --level ...        # position-coupled: which axis? which variant? collides across variants
GOOD  --sink-s3-bucket ...  --compression-zstd-level ...   # each variant registers its own options at its leaf
```

## Consumer symmetry: realizations differ, names rhyme

Two consumers of the same axes should rhyme: the axis _names_ match even when each realizes them with different machinery.

| Axis          | CLI realization (in-process) | service realization (scheduled) |
| ------------- | ---------------------------- | ------------------------------- |
| `format`      | streaming encoder            | encoder stage in job spec       |
| `sink`        | direct file / s3 client      | sink connector in spec          |
| `compression` | inline stream filter         | compression setting in spec     |

Keep the axis vocabulary identical; allow a structural difference only where the realization genuinely differs, and say so out loud. Silent asymmetry is the bug — one side growing an axis the other lacks should be pushed to the shared library or justified explicitly ([applying-the-layout.md](applying-the-layout.md)).

**Least astonishment is the test** — if a reader who has seen `sink/{shared,file,s3}` plus `--sink-s3-bucket` can guess `compression/{shared,zstd}` with `--compression-zstd-level`, the naming is right. Every name a reader must look up instead of derive is a place the symmetry broke; a duller name that fits the pattern beats a sharper one that stands alone.
