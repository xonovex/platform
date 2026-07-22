# Finding the Axes

A variation axis is one decision likely to change; a variant is one answer. Name the question before creating directories.

## Discover

1. List decisions stakeholders may change independently: output format, destination, retry policy.
2. Phrase each as one question: “How is a document encoded?” or “Where is it stored?”
3. Group answers to the same question under one axis.
4. Split a candidate when its parts can change independently; merge candidates that always change together.

For a document processor, `format {json,csv}` × `destination {file,object-store}` × `compression {none,gzip}` is three axes. `json-to-file` and `csv-to-file` are combinations, not variants of one useful axis.

## Tests

- **Swap test** — choosing another value on one axis should not require changing another axis’s variants. If it does, either the boundary is wrong or the exception needs a localized bridge.
- **Question test** — options answering the same question share an axis; a candidate answering two questions must split.
- **Change test** — a likely change to one concept should stay within one axis.

## Do not add an axis when

- There is only one implementation and no demonstrated independent change pressure.
- The proposed shared layer mostly forwards parameters or contains per-variant flags.
- The concern applies to every configuration; that is cross-cutting.
- Two supposed axes always move together; they are one decision wearing two names.

Duplicate a small amount until the stable decision becomes visible. Removing a speculative seam is usually harder than extracting a proven one.
