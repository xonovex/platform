# The Coupling Ladder

Name the worst rung a seam sits on, then push it one rung down. The ranking is total; data coupling is the floor — aim there. The rung names the _shape_ of the dependency; connascence names the _forced co-change_ (see [connascence.md](connascence.md)).

## The ladder (worst to best)

- **Content (worst)** — a module reaches into another's internals (fields, private state, representation). No surface mediates.
- **Common** — modules share global/mutable state; every reader and writer is silently bound through the shared cell.
- **Control** — a caller passes a flag that drives the callee's branching, so it knows the callee's control flow.
- **Stamp** — a caller passes a whole record but the callee uses one field; over-declares its dependency.
- **Data (best)** — plain parameters, no control; the callee depends only on the values it reads.

```go
// Control coupling — BAD: caller drives the callee's branch with a flag.
repo.Save(order, dryRun) // callee: if dryRun { validate } else { persist }
// GOOD: split the branch into named operations.
repo.Validate(order)
repo.Persist(order)
```

## Tie to connascence

- **Content and common** are strongest and least-local — they bind through a representation or shared cell neither call site shows (strong, non-local connascence of value/meaning). Fix first.
- **Control coupling is connascence of meaning across a boundary** — both sides agree what the flag means; splitting into named operations converts it toward connascence of name.
- **Stamp and data are weakest** — a data seam carries only connascence of name/type on the values it names, the most local, compile-time-visible form. Pushing down the ladder and weakening connascence are the same move.

A train wreck like `order.customer().address().city()` is the call-site tell for content coupling — see [law of demeter](law-of-demeter.md).

Back to the overview: [SKILL.md](../SKILL.md).
