# Cross-Cutting Concerns Are Not Axes

Some concerns are present at _every_ point in the design space and encapsulable in _no_ variation point. They are not axes — weave them once at the composition root, never force them into the tree. See [finding-axes.md](finding-axes.md) for what _is_ an axis.

## Three shapes that look alike

- **Variation point (axis)** — partitions the space; each product picks exactly _one_ variant (`format` → `json` _or_ `csv` _or_ `parquet`). Mutually exclusive on one dimension.
- **Cross-tree constraint** — pairwise coupling between _two_ axes, a LOCAL bridge in the dependent leaf (`format/parquet` needs `compression/zstd`'s writer → glue in `format/parquet/zstd.go`). Touches two nodes. See [variation-point-bridges.md](variation-point-bridges.md).
- **Cross-cutting concern** — touches _all_ variants on _every_ axis (the fail-closed encryption gate runs whichever format/sink/compression is selected). Belongs to no node, so it is neither one variant nor one bridge.

## Tangling and scattering

Force a cross-cutting concern into the tree and you get **tangling** (the logic mixed into each variant, so one leaf does two jobs — low cohesion) and/or **scattering** (the same check duplicated across variants, so changing the rule means editing every leaf — connascence of algorithm over N files).

```go
// BAD — policy + logging tangled into every sink, scattered across all of them
func (f file) Write(rec Record) error {
  if !policyOK(rec) { return errDenied }   // copied verbatim into s3, stdout, ...
  log.Info("sink", "kind", "file")
  return f.put(rec)
}
// GOOD — variants do ONLY their one job; the concern is woven once at the root
func (f file) Write(rec Record) error { return f.put(rec) }
```

## Weave once at the composition root

The composition root already knows every selected variant, so it applies the concern generically over all of them — never per plugin. The encryption gate reads each variant's self-declared `Capabilities()` and enforces a required set fail-closed, naming no concrete plugin; adding a `gcs` sink inherits policy, telemetry, logging for free. See **microkernel-pattern-guide**.

```go
func Compose(sel Selection, reg Registry) (Export, error) {
  enc, sink, comp := reg.resolve(sel)
  if err := enforce(sel.Required, capabilities(enc, sink, comp)); err != nil {
    return Export{}, err        // fail closed, once, over the whole selection
  }
  return weave(observe, logRun, Export{enc, sink, comp}), nil
}
```

## Homogeneous vs heterogeneous

- **Homogeneous — identical everywhere → centralize.** Same policy, same log shape: implement once at the root over the neutral surface; zero leaks into a leaf.
- **Heterogeneous — varies per variant → a driven hook.** When the concern genuinely differs (each sink reports a different flush metric), the port declares a narrow hook and the core _drives_ it uniformly; the variant fills only its slice. Core owns the _when/whether_, variant owns the _what_.

```go
type Sink interface {
  Write(Record) error
  Observe() Metrics   // each variant fills its own; the core decides when to collect
}
```

## Canonical cross-cutting concerns

Policy/security gate (the fail-closed archetype), observability/telemetry, logging, caching, validation, persistence, transactions, error handling — each wraps the whole operation; pushing any into one variant scatters it.

**The test** — try adding the concern as a sibling variant. Forces scattering or tangling → it is cross-cutting; weave at the root. Mutually exclusive with the siblings and a product picks one of it → it was a real variant. A cross-cutting concern lives at no node, so [structure-isomorphism.md](structure-isomorphism.md) shows it at no directory.
