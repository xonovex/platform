# Law of Demeter

The principle of least knowledge — "don't talk to strangers" — keeps a call site from depending on the shape of objects it was never handed.

## The rule

- **A method may talk only to its immediate neighbors** — its own members, its parameters, objects it creates inside the method, and its direct components. Nothing further out.
- **It must not reach through a returned object's internals** — once you call `x.thing()` and then call a method on the result, you have started talking to a stranger you never received.
- **State it as least knowledge, not least typing** — the goal is to narrow what each unit must know about the rest of the system, so a change to a distant object's shape stops rippling back to this call site.
- **It is the call-site form of dependency inversion** — ask a neighbor to do the work and let it own its own collaborators, instead of fetching them and steering them yourself.

## The train wreck is a detector

- **`a.b().c().do()` is a train wreck** — each `.` after the first method call hops to a new object the caller never received. The chain reads like a route through someone else's object graph.
- **It signals content coupling** — the worst rung on the [coupling ladder](coupling-ladder.md). The caller is reaching into another module's internals to find the thing it actually wants to act on.
- **It signals strong, non-local connascence** — the caller is now connascent with the type and structure of every intermediate object in the chain, across a distance. See [connascence.md](connascence.md): strong forms must stay local; a train wreck spreads them far.
- **Count the dots, then count the dependencies** — a four-link chain means a change to any of three returned types can break this one line. That degree-three, non-local binding is exactly what the rules of degree and locality tell you to dissolve.

```go
// BAD — reaches through two returned objects: depends on Order AND
// Address internals. A reshape of either Customer() or Address() breaks here.
order.Customer().Address().City()
```

```go
// GOOD — ask the direct neighbor to do the work; depend on one surface.
// The Order owns how it reaches its customer and that customer's address.
order.ShippingCity()
```

## Push the work to the owner

- **Move the behavior to where the data lives** — instead of fetching a stranger and mutating it, tell the object that owns it to perform the change. The caller names an intent; the owner keeps its internals private.
- **A composition root may know many things; a worker may not** — wiring code that assembles a cart legitimately holds many references. Day-to-day call sites must not borrow that reach.

```go
// BAD — the billing routine walks into an Order to read a private total field.
billing.order.total = billing.order.total - discount

// GOOD — the Order self-applies; billing names one neutral operation and
// stays unaware of how the order computes and stores its total.
order.ApplyDiscount(discount)
```

## It is a heuristic, not an absolute

- **Fluent builders are a deliberate exception** — `New().With(x).With(y).Build()` chains because every call returns _the same_ builder, not a fresh stranger. You are talking to one object, not walking a graph.
- **Pure data pipelines are fine** — `xs.Filter(f).Map(g).Reduce(h)` over an immutable collection threads a value through transforms; there is no hidden mutation of someone else's internals to depend on.
- **The target is reaching through to mutate or to depend on a stranger's structure** — apply the rule when a chain crosses object identities and assumes their internal shape, not when it merely reads a self-similar sequence of values.
- **When in doubt, ask what breaks** — if reshaping an intermediate object would break this line, the line knows too much; collapse it behind the direct neighbor's surface.

Back to the overview: [SKILL.md](../SKILL.md).
