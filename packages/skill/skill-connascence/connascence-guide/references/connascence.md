# Connascence

The precise, gradable vocabulary for coupling: name a seam by its kind of forced co-change instead of calling it "tight" or "loose."

## The unit of coupling

- **Two elements are connascent if changing one forces a matching change in the other to stay correct** — that forced co-change _is_ the coupling. Always name the kind precisely; "tight/loose" hides what would actually break.
- **Use it as a diagnostic** — at every seam, ask: if I change this, what else _must_ change in lockstep, and can a reader see why? The answer is the connascence.
- **Every move reduces to three levers** — make the connascence weaker (kind), shorter (distance), or narrower (count). The rest of this file names them.

## Static forms (visible at compile time)

Easiest to find and to weaken; the compiler or a grep can usually point at both ends.

- **Name** — both sides must agree on a name (a field, a key, a symbol). The weakest form; rename tooling catches it.
- **Type** — both sides must agree on a type.
- **Meaning / convention** — both sides must agree on what a value _means_, e.g. `0` means "unlimited". The pact is invisible; a reader sees a number, not the rule.
- **Position** — both sides must agree on an _order_ (positional arguments, tuple slots, column order).
- **Algorithm** — both sides must compute the same way, e.g. a checksum or a hash done identically on each end.

## Dynamic forms (only visible at runtime)

Stronger and harder to find — nothing at the call site reveals them.

- **Execution order** — A must run before B (init before use, open before write).
- **Timing** — correctness depends on _when_, e.g. a race, a timeout, a debounce window.
- **Value** — two values must stay consistent, e.g. a stored total and the parts it sums, or a cache and its source.
- **Identity** — both sides must reference the _same_ instance, not merely an equal one.

## Rate every connascence on three measures

- **Strength** — static is weaker and cheaper than dynamic. A misnamed field fails to compile; a broken execution-order pact fails intermittently in production.
- **Locality** — elements close together (same function, same leaf) may carry strong connascence; elements far apart must carry only weak forms. Distance multiplies cost.
- **Degree** — how many elements are bound by the pact. Fewer is better; a convention shared by two sites is recoverable, one shared by twenty is a refactor wall.

## Two rules that drive every move

- **Rule of degree — convert strong forms into weaker forms.** Replace a positional or conventional pact with a named one; hand a neutral data value across a seam instead of calling into another module's internals.
- **Rule of locality — the farther apart two elements, the weaker their connascence must be.** Strong connascence is fine inside one leaf and forbidden across a boundary. Pull related elements together, or weaken the pact until it can survive the distance.

```go
// Connascence of meaning (static, but non-local): two distant sites must agree
// that timeout 0 means "no limit". A reader at neither end can see the pact.
cache.Get(key, 0)                    // call site
if ttl == 0 { /* unlimited */ }      // far-away handler — silent agreement

// Rule of degree: convert it to connascence of name — the pact is now visible.
cache.Get(key, TTL{Unlimited: true})
```

## Position is why named beats positional

- **Connascence of position is exactly why named, namespaced parameters beat positional conventions** — order-dependence is fragile and non-local; a named option converts it to connascence of name, the weakest form.
- **The further the two ends sit apart, the worse positional gets** — a reorder at the definition silently corrupts every distant call site, with nothing to fail at compile time.

```go
// BAD — connascence of position + meaning across distant sites. Both the call
// and the definition must agree the third bool is "apply discount". Reorder the
// params and every caller is silently wrong.
pricing.Total(cart, true, false, true)

// GOOD — connascence of name. Each field is self-describing; order is free.
pricing.Total(cart, Options{Discount: true, TaxExempt: false, RoundUp: true})
```

## How to use it

- **Grade, then move** — name the seam's strongest connascence and how far it reaches. Then weaken the kind (rule of degree), shorten the distance (rule of locality), or narrow the count (degree).
- **Worked grade** — a billing function that reaches into an `Order`'s private `total` field has connascence of meaning reaching across a module boundary. Fix by both rules at once: have the `Order` expose a neutral computed value the billing function reads, so the agreement is named (degree) and local to one surface (locality), and the billing code stays correct against order shapes it has never seen.
- **Pair it with the ladder** — connascence grades the _pact_; the [coupling ladder](coupling-ladder.md) grades the _channel_ that carries it. Grade both at each seam.
- **Weaken across boundaries, tolerate it within** — pushing a seam to a port (a self-declared value a composition root reads) weakens cross-boundary connascence; isolating what varies onto its own boundary keeps strong connascence local to one unit.

## Named smells that are connascence

Several classic coupling smells are connascence under another name; recognize them, then apply the two rules.

- **Data Clump** — a group of values that always travel together is connascence of position/name repeated across every signature they appear in; reify them into one object (rule of degree) so the pact is named and lives in one place.
- **Feature Envy** — a method more interested in another object's data than its own carries strong connascence to that object across a boundary; move the behavior to the data (rule of locality) so the pact stays local.
- **Shotgun Surgery** — one conceptual change forcing scattered edits across many modules is high-degree, non-local connascence; consolidate the decision behind one owner so degree and distance collapse. (Message chains and inappropriate intimacy are covered by [law-of-demeter.md](law-of-demeter.md) and the [coupling ladder](coupling-ladder.md).)

Back to the overview: [SKILL.md](../SKILL.md).
