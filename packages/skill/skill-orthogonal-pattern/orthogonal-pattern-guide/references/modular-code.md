# Modularity Lives in the Code

Modularity is a property of the dependency edges, not the directory tree. A folder only labels a boundary the interfaces already enforce; move folders without narrowing interfaces and nothing changed. Narrow the seam in code, then let the tree mirror it ([structure-isomorphism.md](structure-isomorphism.md)).

## The rules are the coupling ladder applied to variants and core

**connascence-guide** owns the ladder, connascence, cohesion, and the Law of Demeter; this file maps the rules onto it.

- **Variants never import each other** — no content coupling; each depends on the port, never a sibling.
- **No globals, explicit state** — no common coupling; a shared mutable package global is a back channel that re-couples variants the ports just separated. State travels through arguments.
- **Select a variant, don't pass a behavior flag into the core** — avoid control coupling; a `format string` telling the core which branch to take is the core knowing every leaf. Hand it the chosen `Encoder` and let the variant decide.
- **Neutral data handoff** — aim for data coupling (the weakest, most local rung); one axis hands another a plain value naming _what to apply_, not _who applies it_.

## Depend on the abstraction; the leak test

Core and callers depend on the axis port (dependency inversion); only the composition root imports concrete variants. The core never branches on a variant enum — an encryption gate enforces a self-declared capability (`Encrypted`) and names no concrete sink, staying fail-closed against variants it has never heard of. Dependencies point inward toward stable abstractions: an edge may point from a variant to the port, never from the core to a variant.

```sh
# Leak test: no leaf symbol may appear in core or port code.
grep -rE 'json|csv|parquet|s3' core/ port/ && echo "LEAK: core names a variant"
```

## Narrow, single-purpose ports

A port exposes only what _every_ variant must provide (interface segregation). A method one leaf cannot honor forces no-op stubs; a method from another concern leaks that concern in. When stubs appear, carve off the extra methods into a smaller port.

```go
// BAD — fat port: csv must stub Schema(); a compression concern leaks in
type Encoder interface {
    Encode(r *Record) ([]byte, error)
    Schema() (Schema, error)    // only parquet has a schema; others return nil,nil
    Compress() Compressor       // belongs to the compression axis
}
// GOOD — narrow: only what every encoder provides
type Encoder interface {
    Encode(r *Record) ([]byte, error)
    ContentType() string
}
```

## Pass data across seams, not calls into internals

Hand a neutral data VALUE between axes (an encoded `Record` plus content-type and metadata), not a method call into the other's guts — that is what makes the two axes independently swappable and testable. Honor the Law of Demeter at the call site: a leaf reaching through another's returned internals (`a.b.getC().do()`) bypasses the bridge and rebuilds content coupling one dot at a time. See [variation-point-bridges.md](variation-point-bridges.md).

## Explicit state and direct dependencies

- **No hidden singletons or runtime-mutated package globals** — pass context explicitly; prefer module-level functions over classes (full discipline in fp-guide).
- **No re-export shims** — import a symbol from its owner, not through a barrel that couples two axes and hides the real edge (`sink/shared` re-exporting `format.Record`). Delete dead/deprecated paths instead of wrapping them.
- **Small, focused files** — a leaf is its variant and nothing else; the shared core is the contract and nothing else.

## False modularity and testability

- **A fan-out tree proves nothing on its own** — neat `format/`, `sink/`, `compression/` folders are false modularity (a distributed monolith) if leaves still import siblings or share mutable state. Enforce with import-cycle and architecture-test fitness functions that fail the build when a leaf names a sibling or the core names a leaf, plus single data ownership per axis. See [boundary-alignment.md](boundary-alignment.md).
- **Testability is the proof** — if a variant unit-tests against a fake port with no globals and no full-system wiring, the dependency was inverted and state was explicit. Needing the whole graph (real registry, real S3, real zstd) to exercise one rule means the seam is too wide — narrow the port, not the test.

```go
func TestPolicyDeniesUnencryptedSink(t *testing.T) {
    sink := fakeSink{capabilities: Plain}              // a fake satisfies the port
    if err := Enforce(sink, &Record{}); err == nil {   // no registry, S3, or zstd writer
        t.Fatal("expected fail-closed policy rejection")
    }
}
```

See **microkernel-pattern-guide** for the registry composition root.
