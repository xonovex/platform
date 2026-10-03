# Applying the Layout

Map each proven axis to one directory. Put the universal contract in `shared/` and use bare sibling leaves for variants.

```text
document-processing/
  format/{shared,json,csv}/
  destination/{shared,file,object-store}/
  compression/{shared,none,gzip}/
  compose
```

## Incremental migration

1. Name the independent decisions without moving files.
2. Assign every existing file to one axis or mark it cross-cutting.
3. Create one axis directory and move its variant-specific code into symmetric leaves.
4. Move only universally shared contracts to that axis's `shared/`.
5. Localize genuine two-axis exceptions as bridges.
6. Repeat one axis at a time, keeping the system working after each move.
7. Remove catch-all `utils` or layer directories after their contents have clear owners.

If a proposed axis produces wrappers, ignored options, or changes that still span multiple leaves, reverse it: inline the shared code into the leaves, remove the seam, and wait for clearer change pressure.

The tree communicates the model but does not enforce it. Port design belongs to **hexagonal-pattern-guide**, registration to **microkernel-pattern-guide**, and coupling enforcement analysis to **connascence-guide**.
