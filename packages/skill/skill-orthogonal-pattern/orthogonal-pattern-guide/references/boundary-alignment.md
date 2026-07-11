# Boundary Alignment

An axis must be cut on the right boundary, not just laid out tidily — a well-named axis on a wrong seam still leaks. This file answers _where_ the boundary goes, then how to make it real. For _which_ decisions are axes, see [finding-axes.md](finding-axes.md).

- **Placement before layout** — a clean port does nothing if the seam runs through the middle of one concept; the connascence just moves inside the leaf and surfaces as cross-axis edits later. The test is "does a change to one concept stay on one side?", not "is the directory pretty?". Coupling vocabulary is owned by **connascence-guide**.

## Align to a domain seam, not a layer

An axis should coincide with a bounded context — a region with its own consistent vocabulary (ubiquitous language) and one reason to change (both owned by **ddd-guide**). `format` is a seam because "how is a record encoded?" has its own vocabulary (`Columnar`, `RowDelimited`) that means nothing to `compression`. Cut on a technical layer or storage table instead and one concept splits across two owners.

```
BAD   seam on the storage/layer cut: one concept, two owners
  format/           decides the column schema
  config/sink_rules row also decides the column schema, keyed by format
        → change "parquet is columnar" → edit format AND the rules table
GOOD  seam on the domain concept: encoding is one context
  format/{json,csv,parquet}  each leaf self-declares its capabilities
        → change "parquet is columnar" → edit parquet leaf only
```

## One owner per axis (Conway's law)

Structure mirrors the communication structure of the teams. An axis owned by no one, or split across two teams, drifts back into coupling.

- **One owner per axis** — that team owns the port, the value handed across the seam, and every leaf. Shared ownership means two change cadences blurring one seam.
- **No-owner axis rots** — it accretes special cases from whoever touches it last; self-declared capabilities stop being trustworthy.
- **The inverse maneuver is a deliberate lever** — to reach a target architecture, give the three axes three owners _on purpose_ so the seams you want are the seams that get defended.

## Enforce in code, not folders

A boundary exists only when something fails the build when it is crossed; a directory tree alone is false modularity. A fitness function asserts the edges: the encryption gate imports no concrete leaf, `shared/` imports no variant, no two leaves import each other. Data ownership backs it — one axis owns its types; another reaches them only through the neutral `Record`.

```go
// fitness function: the encryption gate must stay sink-agnostic and fail-closed
func TestPolicyImportsNoLeaf(t *testing.T) {
    deps := importsOf(t, "internal/policy")
    for _, leaf := range []string{"sink/s3", "sink/file", "format/parquet"} {
        if deps.Has(leaf) {
            t.Fatalf("policy imports %s — boundary breached, gate now knows a variant", leaf)
        }
    }
}
```

If the test is green only because no one wrote it, the boundary is folklore. See [modular-code.md](modular-code.md) for the dependency rule, [applying-the-layout.md](applying-the-layout.md) for placing the tree.

## The alignment test

A boundary is correctly placed when a change to one domain concept stays inside one axis and one owner. When it fans out across two axes or two teams, the seam cuts through a concept that wants to be whole (or two concepts share an owner that wants to be split). Don't patch — re-cut: merge two axes that always change together, or split one whose halves move independently (reversal recipe in [applying-the-layout.md](applying-the-layout.md)).

```
ALIGNMENT TEST
  pick a likely change to ONE concept ("s3 sink must encrypt uploads")
  → stays in one axis + one owner?   yes → aligned
  → touches sink AND compression?     no  → seam misplaced; re-cut
```
