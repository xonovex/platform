# Connascence

Two elements are connascent if changing one forces a matching change in the other. Name the kind and rate it instead of saying "tight/loose"; at each seam ask what else _must_ change in lockstep and whether a reader can see why.

## Static forms (compile-time visible), weakest to strongest

- **Name**: both sides agree on a name (field, key, symbol). Weakest; rename tooling catches it.
- **Type**: both sides agree on a type.
- **Meaning / convention**: both sides agree what a value _means_ (e.g. `0` means "unlimited"). The pact is invisible.
- **Position**: both sides agree on an _order_ (positional args, tuple slots, column order).
- **Algorithm**: both sides compute identically (a checksum/hash done the same on each end).

## Dynamic forms (runtime-only, stronger, harder to find)

- **Execution order**: A must run before B (init before use).
- **Timing**: correctness depends on _when_ (race, timeout, debounce).
- **Value**: two values must stay consistent (a stored total and its parts; a cache and its source).
- **Identity**: both sides must reference the _same_ instance, not an equal one.

## Rate on three measures

- **Strength**: static is weaker/cheaper than dynamic (fails to compile vs fails intermittently in prod).
- **Locality**: close elements may carry strong connascence; distant elements must carry only weak forms.
- **Degree**: how many elements are bound; fewer is better.

## Two rules

- **Rule of degree**: convert strong forms into weaker ones (positional/conventional pact → named).
- **Rule of locality**: the farther apart two elements, the weaker their connascence must be. Strong connascence is fine inside one leaf, forbidden across a boundary.

```go
// Connascence of meaning (non-local): distant sites must agree 0 means "no limit".
cache.Get(key, 0)                    // call site
if ttl == 0 { /* unlimited */ }      // far-away handler: silent pact
// Rule of degree → connascence of name makes the pact visible.
```

Connascence grades the _pact_; the [coupling ladder](coupling-ladder.md) grades the _channel_ carrying it: grade both.

## Classic smells that are connascence

- **Data Clump**: values that always travel together = repeated position/name connascence across signatures; reify into one object.
- **Feature Envy**: a method more interested in another object's data carries strong connascence across a boundary; move behavior to the data (rule of locality).
- **Shotgun Surgery**: one change forcing scattered edits = high-degree non-local connascence; consolidate behind one owner.

Back to the overview: [SKILL.md](../SKILL.md).
