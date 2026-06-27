# Cross-Cutting Concerns Are Not Axes

The honest limit of this skill: some concerns are present at _every_ point in the design space and encapsulable in _no_ variation point. They are not axes — do not force them into the tree. Weave them once at the composition root instead. See [finding-axes.md](finding-axes.md) for what _is_ an axis, and [SKILL.md](../SKILL.md) for the map.

## Three things that look alike but are not

- **A variation point (axis)** — partitions the design space; each product picks exactly _one_ variant. `format` answers "how is the record encoded?" and a config picks `json` _or_ `csv` _or_ `parquet`. Mutually exclusive choices on one dimension.
- **A cross-tree constraint** — pairwise coupling between _two_ variation points, a LOCAL bridge living in the dependent leaf. `format/parquet` needs the writer that `compression/zstd` builds: glue placed in `format/parquet/zstd.go`, touching two nodes, not all of them. See [variation-point-bridges.md](variation-point-bridges.md).
- **A cross-cutting concern** — touches _all_ variants on _every_ axis at once. The fail-closed encryption gate must run whichever format, sink, and compression variants are selected. It belongs to no single node, so it cannot be one variant and cannot be one bridge.

A cross-tree constraint is connascence between two leaves you can localize; a cross-cutting concern is connascence between the concern and _every_ leaf, which is why no leaf can own it. For the coupling vocabulary used here, see **connascence-guide**.

## Tangling and scattering: the two failure modes

Force a cross-cutting concern into the tree and you get one or both:

- **Tangling** — the cross-cutting logic gets mixed _into_ each variant, so one leaf does two unrelated jobs (write the bytes _and_ enforce policy _and_ emit telemetry). Low cohesion; the variant can no longer be read as one thing.
- **Scattering** — the _same_ check is duplicated _across_ variants, so changing the rule means editing every leaf and remembering each one. Connascence of algorithm spread over N files — change one, the others silently drift.

```go
// BAD — policy + logging tangled into every sink, scattered across all of them
func (f file) Write(rec Record) error {
  if !policyOK(rec) { return errDenied }   // copied verbatim into s3, stdout, ...
  log.Info("sink", "kind", "file")         // copied again, edited everywhere
  return f.put(rec)
}
func (s s3) Write(rec Record) error {
  if !policyOK(rec) { return errDenied }   // same four lines, fourth copy
  log.Info("sink", "kind", "s3")
  return s.put(rec)
}
```

```go
// GOOD — variants do ONLY their one job; the concern is woven once at the root
func (f file) Write(rec Record) error { return f.put(rec) }
func (s s3) Write(rec Record) error { return s.put(rec) }
```

## Weave the concern once at the composition root

The composition root is the _weaving point_: the one place that already knows every selected variant, so it can apply the concern generically over all of them.

- **Weave generically, never per plugin** — the encryption gate reads each variant's self-declared `Capabilities()` and enforces a required set without naming any concrete plugin. It is the archetype cross-cutting concern: woven once over every variant, fail-closed by default. See **microkernel-pattern-guide**.
- **The wrapper sees all axes** — wrap the assembled export, not the leaves. Adding a `gcs` sink inherits policy, telemetry, and logging for free, because none of them live in a leaf.

```go
// registry/root: weave the concern over WHATEVER variants were selected
func Compose(sel Selection, reg Registry) (Export, error) {
  enc, sink, comp := reg.resolve(sel)
  if err := enforce(sel.Required, capabilities(enc, sink, comp)); err != nil {
    return Export{}, err        // fail closed, once, over the whole selection
  }
  return weave(observe, logRun, Export{enc, sink, comp}), nil
}
```

## Homogeneous vs heterogeneous

- **Homogeneous — identical everywhere → centralize.** The behavior does not vary by variant: the same policy decision, the same log shape. Implement it once at the root over the neutral surface (the encoded `Record` value, the declared capabilities), and let zero of it leak into a leaf.
- **Heterogeneous — varies per variant → a per-variant hook the core still drives.** When the concern genuinely differs (each sink reports a different flush metric), the core declares a narrow hook on the port and _drives_ it uniformly; the variant fills in only its own slice. The core still owns the _when_ and the _whether_; the variant owns only its _what_. That keeps it a driven hook, not scattered logic.

```go
// Heterogeneous concern: the port declares the hook, the core calls it for ALL variants
type Sink interface {
  Write(Record) error
  Observe() Metrics   // each variant fills its own; the core decides when to collect
}
```

## Canonical cross-cutting concerns

Recognize these by their reach across every axis — none is a variant or a bridge:

- **Policy / security gate** — the fail-closed archetype; enforced once over all selected variants.
- **Observability / telemetry** and **logging** — emitted around the woven export, not inside leaves.
- **Caching**, **validation**, **persistence**, **transactions**, **error handling** — each wraps the whole operation; pushing any into a single variant scatters it into the rest.

## Why it is invisible in the tree, and the test

- **It lives at no single node, so it appears at none.** The directory tree is isomorphic to the _axes_; a cross-cutting concern is orthogonal to all of them, so it has no home directory. Expect to find it at the weaving point, not as a sibling leaf — see [structure-isomorphism.md](structure-isomorphism.md).
- **The test** — try adding the concern as a sibling variant. If doing so forces scattering (the same check copied into siblings) or tangling (a leaf now doing two jobs), it is cross-cutting: pull it back out and weave it at the root. If instead it is mutually exclusive with the existing siblings and a product picks one of it, it was a real variant after all.
