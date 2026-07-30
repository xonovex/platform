# Law of Demeter

Least knowledge, "don't talk to strangers." A method may talk only to its immediate neighbors: its own members, its parameters, objects it creates inside itself, and its direct components. It must not reach through a _returned_ object's internals.

## The train wreck is a detector

`a.b().c().do()`: each `.` after the first method call hops to a new object the caller never received. It signals **content coupling** (the worst [coupling-ladder](coupling-ladder.md) rung) and strong, non-local connascence with every intermediate type.

```go
// BAD: reaches through two returned objects; depends on Order AND Address internals.
order.Customer().Address().City()
// GOOD: ask the direct neighbor; depend on one surface.
order.ShippingCity()
```

## Exceptions: it is a heuristic, not an absolute

- **Pure data pipelines**: `xs.Filter(f).Map(g).Reduce(h)` over an immutable collection threads a value through transforms; no hidden mutation of another's internals.
- The target is reaching through to _mutate or depend on a stranger's structure_, not reading a self-similar sequence of values. When in doubt: if reshaping an intermediate object would break the line, it knows too much.
- A composition root may hold many references; day-to-day call sites may not borrow that reach.

Back to the overview: [SKILL.md](../SKILL.md).
