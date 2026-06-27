# Naming and Symmetry

Parallel concepts deserve parallel names. When a tree's siblings are named consistently, a reader who learns one of them can guess the rest — that predictability is the whole payoff of a clean decomposition. A directory names a _variation point_ (an axis: a design decision likely to change); a leaf names a _variant_ (one value of that decision). This file is about the names themselves; for the tree they hang on see [structure-isomorphism.md](structure-isomorphism.md), and for what lives inside each node see [commonality-variability.md](commonality-variability.md).

## Sibling concepts get sibling names

- **Same part of speech, same level of abstraction** — siblings under one parent must answer the same kind of question at the same altitude. A variation point names the _question_, so it is a noun: `format`, `sink`, `compression`. A directory named for a mechanism or verb (`encode`, `upload`) reads as a different category and breaks the parallel.
- **One odd sibling poisons the set** — the moment `encode/` sits beside `sink/` and `compression/`, a reader can no longer tell whether `encode` is a variation point, a tool, or a grab-bag. Rename it to the variation point it actually realizes (`format`) so every sibling is the same kind of thing.

```
BAD  internal/{encode, sink, compression}/   # 'encode' is a verb, the rest are axes
GOOD internal/{format, sink, compression}/   # three nouns, three questions, one altitude
```

## The axis dir supplies context — keep the leaf bare

- **The path already says it** — `sink/s3` reads as "the s3 variant of sink." The variation-point directory supplies the context, so the variant leaf stays bare. Suffixing the leaf to `sink/s3uploader` repeats the parent; the qualified path makes the suffix redundant noise. Name the leaf for the _variant_, nothing more.
- **A bare leaf composes** — `format/json`, `compression/zstd`, and `sink/s3` all read uniformly; tool-tagged leaves like `jsonformat` force the reader to re-parse each one.

| BAD (leaf repeats axis)       | GOOD (bare leaf)   |
| ----------------------------- | ------------------ |
| `sink/s3uploader`             | `sink/s3`          |
| `format/jsonformat`           | `format/json`      |
| `sink/stdoutsink`             | `sink/stdout`      |
| `compression/gzipcompression` | `compression/gzip` |
| `sink/filewriter`             | `sink/file`        |

## Name the axis, not the tool you happened to pick

- **The variation point is the noun; the variant is the leaf** — a directory named for an implementation (`s3uploader/`) or a utility blob (`exportutil/`, `writer/`) instead of its variation point is the smell. `s3uploader` → `sink/s3`; `writer` or `exportutil` → the axis's `shared/`. The dir is the question; the tool is one of its answers.
- **Homeless helpers belong inside their axis** — a top-level `exportutil/` sitting beside the axes means a variation point's own helpers leaked out. Push them into that axis's shared node (`sink/shared/`), so the axis owns its commonality.

```
BAD  internal/export/{s3uploader,file}/   internal/exportutil/   internal/writer/
GOOD internal/sink/{shared,file,s3}/   internal/format/{shared,...}/
```

## Options are namespaced by axis + variant

- **The flag name encodes the path** — a per-variant option carries its variation point and variant as a prefix, so `--sink-s3-bucket` and `--compression-zstd-level` read as the directory they configure. The flag namespace is isomorphic to the tree.
- **Named beats positional — it weakens the connascence** — namespaced named options exist to kill order-dependence. Positional or order-by-convention flags impose connascence of position: caller and parser must agree on order, and a reordering silently breaks them. A name (`--sink-s3-bucket`) is connascence of name instead — the weaker, more local form, and the rule of degree says prefer it. The axis+variant prefix also makes each name unique across variants, so `--level` for two variants never collides. See connascence and the coupling/cohesion ladder in **connascence-guide**.
- **Adding a variant adds a leaf, not an edit** — a new sink variant adds `sink/<variant>/` plus its own `--sink-<variant>-*` options. It does not touch a central god-parser that knows every flag. Each variant owns its own options the way each leaf owns its own code.

```
BAD   --bucket ...  --level ...        # position-coupled: which axis? which variant? collides across variants
      func parseAllFlags() { /* one switch that knows every variant */ }
GOOD  --sink-s3-bucket ...  --compression-zstd-level ...
      each variant registers its own --<axis>-<variant>-* options at its leaf
```

## Consumer symmetry: realizations differ, names rhyme

- **Two consumers of the same axes should rhyme** — when a CLI export command (writing in-process) and a scheduled export service (emitting a job spec) both realize the same variation points, their structures mirror each other. The axis _names_ match even though each side realizes them with different machinery.

| Axis          | CLI realization (in-process) | service realization (scheduled) |
| ------------- | ---------------------------- | ------------------------------- |
| `format`      | streaming encoder            | encoder stage in job spec       |
| `sink`        | direct file / s3 client      | sink connector in spec          |
| `compression` | inline stream filter         | compression setting in spec     |

- **Match the names, call out the gaps** — keep the axis vocabulary identical across consumers; allow a structural difference only where the realization genuinely differs, and say so out loud (a comment, a doc line). Silent asymmetry is the bug — one side growing a variation point the other lacks should be either pushed down to the shared library or justified explicitly. See [applying-the-layout.md](applying-the-layout.md) for keeping two consumers aligned.

## Least astonishment is the test

- **Learn one axis, predict the rest** — if a reader who has seen `sink/{shared,file,s3}` plus `--sink-s3-bucket` can correctly guess that compression lives at `compression/{shared,zstd}` with `--compression-zstd-level`, the naming is right. Every name a reader has to look up instead of derive is a place the symmetry broke.
- **Consistency beats cleverness** — a duller name that fits the pattern beats a sharper one that stands alone; the value is in the set being uniform, not in any single name. Back to [SKILL.md](../SKILL.md).
