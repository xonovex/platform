# Modularity Lives in the Code

Modularity is a property of the code's dependency edges, not of the directory tree. A folder only labels a boundary the interfaces already enforce; move folders without narrowing interfaces and you have changed nothing.

## Code first, directories second

- **A variant is a narrow public surface over hidden internals** — the boundary is real only when the exported surface is small and the rest is unreachable. The directory reflects that boundary; it does not create it.
- **Folder moves without interface narrowing are cosmetic** — sorting files into `format/` and `sink/` while every type still reaches into every other type buys nothing. Narrow the seam in the code, then let the tree mirror it; see [structure-isomorphism.md](structure-isomorphism.md).

## Map the rules onto the coupling ladder

The skill's modularity rules are the coupling ladder applied to variants and the core. This file maps the rules; **connascence-guide** owns the ladder, connascence, cohesion, and the Law of Demeter.

- **Variants never import each other — no content coupling** — a leaf that reaches into another leaf's internals binds to symbols the other never meant to publish. Each variant depends on the port, never on a sibling.
- **No globals, explicit state — no common coupling** — a shared mutable package global is a back channel that re-couples variants the ports just separated. State travels through arguments.
- **Select a variant, do not pass a behavior flag into the core — avoid control coupling** — a `format string` that tells the core which branch to take is the core knowing every leaf. Hand it the chosen `Encoder` instead and let the variant decide.
- **Neutral data handoff — aim for data coupling** — one axis hands another a plain value naming _what to apply_, not _who applies it_. Data coupling is the weakest, most local rung; it is the target for every seam.

## Depend on the abstraction, not the implementation

- **Core and callers depend on the axis port; only the composition root depends on variants** — this is dependency inversion. The core knows `Encoder`/`Sink`; the registry is the one place that imports `json`, `csv`, `parquet` and binds each to its port.
- **The core is variant-agnostic** — it never imports a concrete variant and never branches on a variant enum. An encryption gate enforces a self-declared capability (`Encrypted`) and names no concrete sink, so it stays fail-closed against variants it has never heard of.

```go
// BAD — core imports concrete leaves and branches on a sink string (control coupling)
import ("sink/file"; "sink/s3")
func Enforce(sink string, r *Record) error {
    if sink == "s3" { return s3.Check(r) }  // core knows every leaf
    return file.Check(r)                     // a new leaf edits the core
}

// GOOD — core depends only on the port; policy names no plugin
func Enforce(sink Sink, r *Record) error {
    if !sink.Capabilities().Has(Encrypted) {  // self-declared capability
        return ErrPolicy                       // fail-closed, plugin-agnostic
    }
    return nil
}
```

## The Dependency Rule and the leak test

- **Dependencies point inward toward stable abstractions** — the core and the port are the stable inner ring; variants are the volatile outer ring. An edge may point from a variant to the port, never from the core to a variant.
- **An outer name must never appear in inner code** — grep the core for any variant symbol; a hit is a leak that re-couples the core to a leaf it should never name.

```sh
# Leak test: no leaf symbol may appear in core or port code.
grep -rE 'json|csv|parquet|s3' core/ port/ && echo "LEAK: core names a variant"
```

## Narrow, single-purpose ports

- **A port exposes only what EVERY variant must provide** — interface segregation. A method one leaf cannot honor forces the rest to write no-op stubs, and a method from another concern leaks that concern into the port.
- **Split a fat port into the minimum each consumer needs** — when stubs appear, the port is too wide; carve off the extra methods into a smaller port the relevant leaves implement.

```go
// BAD — fat port: csv must stub a schema method; a compression concern leaks in
type Encoder interface {
    Encode(r *Record) ([]byte, error)
    Schema() (Schema, error)    // only parquet has a schema; others return nil,nil
    Compress() Compressor       // belongs to the compression axis, not format
}
// GOOD — narrow port: only what every encoder must provide
type Encoder interface {
    Encode(r *Record) ([]byte, error)
    ContentType() string
}
```

## Pass data across seams, don't call into internals

- **Hand a neutral data VALUE between axes, not a method call into the other's guts** — an encoder returns a `[]byte` (the encoded `Record` plus its content-type and metadata); each compressor consumes it its own way. Neutral data is what makes the two axes independently swappable and testable.
- **Neutral data is the seam** — because the value names _what to apply_, not _who applies it_, any encoder pairs with any compressor with no bespoke glue. See [variation-point-bridges.md](variation-point-bridges.md) for placing that handoff, and [commonality-variability.md](commonality-variability.md) for which axis owns the type.
- **Honor the Law of Demeter at the call site** — a leaf must not reach through another leaf's returned internals (no `a.b.getC().do()`); that walk bypasses the bridge and rebuilds content coupling one dot at a time. **connascence-guide** defines the rule.

## Explicit state, no globals

- **Pass context explicitly; prefer module-level functions over classes; keep no hidden singletons or runtime-mutated package globals** — pure functions where practical. State that travels through arguments is state a test can supply and a second axis can swap.
- A mutable package global is a back channel that re-entangles modules the ports just separated. For the full functional discipline, see fp-guide rather than restating it here.

## Direct dependencies, no re-export shims

- **Import a symbol from its owner, not through a barrel** — a module that re-exports another's types couples the two and hides the real edge: a caller reading `sink` cannot tell the type belongs to `format`.
- **Delete dead and deprecated paths instead of wrapping them** — a back-compat shim is a second public surface for the same thing; remove it so each concept has exactly one import edge.

```
BAD  sink/shared re-exports format.Record
       callers import it "through" sink — two axes coupled, real edge hidden
GOOD  callers import format.Record from its owner, the format axis
```

## Small, focused files

- **One cohesive concern per file** — a leaf is its variant and nothing else; the shared core is the contract and nothing else. The public API is the smallest surface that lets a new variant plug in.

## False modularity is a distributed monolith

- **A fan-out tree proves nothing on its own** — neat `format/`, `sink/`, `compression/` folders are false modularity if the leaves still import siblings or share mutable state. The shape looks decomposed while the edges stay fully coupled: a distributed monolith, where one variant cannot change without the others.
- **Enforce with edges, not layout** — assert the boundary in code: import-cycle and architecture-test fitness functions that fail the build when a leaf names a sibling or the core names a leaf, plus single data ownership per axis. Directory layout documents the boundary; the tests are what hold it. See [applying-the-layout.md](applying-the-layout.md) for the layout and [boundary-alignment.md](boundary-alignment.md) for enforcing single ownership.

## Testability is the proof of modularity

- **If a variant unit-tests against a fake port, with no globals and no full-system wiring, it is modular** — that the test can stand in a fake is the evidence the dependency was inverted and the state was explicit.
- **If a test needs the whole graph stood up, the code is entangled** — needing a real registry, a real S3 bucket, and a real zstd writer to exercise one policy rule means the seam is too wide or the dependency points the wrong way. Narrow the port or invert the dependency, not the test.

```go
// Modular: the rule is exercised against a stand-in port, no graph required.
func TestPolicyDeniesUnencryptedSink(t *testing.T) {
    sink := fakeSink{capabilities: Plain}              // a fake satisfies the port
    if err := Enforce(sink, &Record{}); err == nil {   // no registry, S3, or zstd writer
        t.Fatal("expected fail-closed policy rejection")
    }
}
```

See also **microkernel-pattern-guide** for the registry composition root. Back to the overview: [SKILL.md](../SKILL.md).
