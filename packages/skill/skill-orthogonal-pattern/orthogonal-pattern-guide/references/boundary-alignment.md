# Boundary Alignment

A variation point — the design decision likely to change — must be cut on the right boundary, not just laid out tidily. A well-named axis on a wrong seam still leaks: the question this file answers is _where_ the boundary goes, then how to make it real. For _which_ decisions are variation points at all, see [finding-axes.md](finding-axes.md).

## Where, not just how

- **A boundary is a placement decision before it is a layout decision** — naming an axis `compression` and giving it a clean port does nothing if the seam runs through the middle of one domain concept. The connascence just moves inside the leaf and surfaces as cross-axis edits later.
- **Align with a reason to change, not a noun you happened to type** — the test is "does a change to one concept stay on one side?", not "is the directory pretty?". Coupling vocabulary (connascence, the coupling ladder, the cohesion ladder, Law of Demeter) is owned by **connascence-guide**; this file uses it without redefining it.

## Align to a domain seam, not a layer

A variation point should coincide with a bounded context — a region with its own consistent vocabulary (its ubiquitous language) and one reason to change. Cut on a technical layer or a storage table instead and you split a single concept across two owners.

- **One bounded context, one variation point** — `format` is a seam because "how is a record encoded?" has its own vocabulary (`Columnar`, `RowDelimited`) that means nothing to `compression`. The seam follows the concept, not the call stack.
- **A seam through a concept leaks** — if encoding rules live half in `format` and half in a `compression` table keyed by format, a single encoding change fans out across both. That fan-out is connascence crossing the boundary — the concept was split, not separated.

```
BAD   seam on the storage/layer cut: one concept, two owners
  format/           decides the column schema
  config/sink_rules row also decides the column schema, keyed by format
        → change "parquet is columnar" → edit format AND the rules table

GOOD  seam on the domain concept: encoding is one context
  format/{json,csv,parquet}  each leaf self-declares its capabilities
        → change "parquet is columnar" → edit parquet leaf only
```

This is what makes "one reason to change" concrete: the boundary is right when the change to a concept stays inside the context that owns its words. See [finding-axes.md](finding-axes.md) for deriving the concept; [structure-isomorphism.md](structure-isomorphism.md) for mirroring it on disk.

## Give each variation point a single owner (Conway's law)

System structure tends to mirror the communication structure of the teams that build it. A variation point owned by no one, or split across two teams, drifts back into coupling no matter how clean its port started.

- **One owner per axis** — the team that owns `compression` owns its port, the `[]byte` value it hands across the seam, and every leaf under it. Shared ownership means two change cadences pushing one seam in different directions until it blurs.
- **No-owner axis rots** — a `format` seam nobody owns accretes special cases from whoever touches it last; the self-declared capabilities stop being trustworthy because no single party defends them.
- **The inverse maneuver is a deliberate org lever** — when you want a target architecture, shape the teams to it: give the three axes (`format`, `sink`, `compression`) three owners _on purpose_ so the seams you want are the seams that get defended. Use it intentionally, not by accident.

## Enforce in code, not in folders

A boundary exists only when something fails the build when it is crossed. A directory tree alone is false modularity — a distributed monolith wearing a tidy layout. Folders document intent; fitness functions enforce it. See [modular-code.md](modular-code.md) for the dependency rule the test guards, and [applying-the-layout.md](applying-the-layout.md) for placing the tree.

- **Architecture / import-cycle tests are the enforcement** — a fitness function asserts the dependency edges: the encryption gate imports no concrete leaf, `shared/` imports no variant, no two leaves import each other.
- **Data ownership and visibility back it** — one axis owns its types; another reaches them only through the neutral encoded-`Record` value, never by importing across the seam.

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

If the test is green only because no one wrote it, the boundary is folklore. Write the test; the seam is whatever the test defends.

## The alignment test

A boundary is correctly placed when a change to one domain concept stays inside one variation point and one owner.

- **Right boundary** — "parquet now writes columnar" edits the `parquet` leaf and nothing else; one concept, one axis, one team.
- **Misplaced boundary** — a single change fans out across two axes or two teams. That fan-out is the symptom; the seam is cutting through a concept that wants to be whole, or two concepts share an owner that wants to be split.
- **Re-cut, do not patch** — when the alignment test fails, move the seam: merge two axes that always change together, or split one axis whose halves move independently. Run the reversal recipe in [applying-the-layout.md](applying-the-layout.md) — re-cutting the boundary is a refactor of the tree, not a new sibling dir to hold the glue.

```
ALIGNMENT TEST
  pick a likely change to ONE concept ("s3 sink must encrypt uploads")
  → does it stay in one axis + one owner?     yes → boundary aligned
  → does it touch sink AND compression?        no  → seam misplaced; re-cut
```

Back to the overview: [SKILL.md](../SKILL.md).
