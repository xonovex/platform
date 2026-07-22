# Cleanup Inventories

Produce evidence-linked, read-only inventories for focused cleanup work. Do not edit,
delete, extract, or scaffold code during the audit.

## Barrel exports

- Find subdirectory `index.ts` and `index.js` files; exclude a package's public root
  entry point unless the request explicitly includes it.
- Confirm each candidate contains only re-exports.
- Map every indirect import to the concrete source file and report the deletion and
  import-migration surface.

## Redundant comments

- Detect the language's comment syntax and separate directives from prose.
- Preserve functional directives and non-obvious rationale, invariants, workarounds,
  and caveats.
- Flag comments that restate well-named code, narrate provenance or future work, or
  leave old code commented out.

## Shared extraction candidates

- Find repeated functions, components, hooks, middleware, types, and constants by
  semantic behavior rather than text alone.
- Group identical code, equivalent logic, and shared interfaces separately.
- Rank candidates by occurrence count, complexity, and cross-package reach; report
  coupling and dependency-direction risks before recommending extraction.

## TODO and FIXME inventory

- Scan recursively for `TODO`, `FIXME`, and project-defined markers.
- Normalize messages, group them by intent and affected area, and preserve exact
  file-and-line evidence.
- Grade actionable clusters by impact and likelihood. Do not infer that an old marker
  is dead without checking the surrounding code and repository history.
