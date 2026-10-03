# prototypes: Prototype / Instance / Override Inheritance

## Guideline

Let any object act as a prototype for another object of the same type, where the instance inherits every property and stores only its _overrides_: tracked by a per-property bitmask, with sub-object collections tracked as inherited / instantiated / removed, so a change to the prototype propagates to all instances except the values they deliberately changed, and a save records only the deltas.

## Rationale

Editors need "edit the template, all copies update, but keep the local tweaks" (prefabs, presets, blueprints). Instance-level prototype inheritance where _any_ object can be a prototype beats class inheritance: it applies to any data type, lives in data not code, needs no class hierarchy, and stays decidable (no computed-property expressions, so no surprise recomputation). Storing only overrides is the payoff: files shrink to the diff from the prototype, merges stay clean, an un-overridden property tracks the prototype automatically. A per-property override bitmask makes "inherited or overridden?" O(1); resolving a property walks the chain until someone overrides it.

## How to Apply

1. Make prototype-of a relationship between two objects of the same type (a reference from instance to prototype), not a special "prefab" object kind.
2. Resolve a property by checking the instance's override bitmask: if the bit is set, use the instance's value; otherwise read through to the prototype, recursively up the whole chain.
3. Setting a property sets its override bit; "revert to prototype" clears the bit and drops the local value.
4. For sub-object collections (sets), track three states per child: _inherited_ (comes from the prototype untouched), _instantiated_ (a local overridable copy, itself a child prototype), and _removed_ (suppressed from the inherited set). Represent edits as add/remove/instantiate ops, not a wholesale value override.
5. Propagate a prototype edit to instances at resolve time (instances read through), so no push/copy step is needed.

## Gotchas

- Instantiating (overriding) a sub-object usually gives it a new id, so any reference pointing at the _prototype's_ sub-object now misses: resolve references through the override (redirect to the instantiated copy) instead of storing the new id everywhere.
- Editing a deeply nested inherited sub-object requires instantiating every parent in the chain first (you can't override what you haven't materialized): expose a one-click "instantiate ancestors" or this becomes tedious "drilling down."
- Resist computed/derived overrides (`hp = proto.hp * 1.5`): they reintroduce evaluation order, cycles, and a mini-language into the data model. Keep overrides concrete values.
- A removed inherited child must stay suppressed even after the prototype changes; store removal as an explicit op, not as "absent," or a prototype edit silently re-adds it.
- Reverting an override must clear the bit _and_ the stored value, or stale local data leaks back when the bit is later re-set.

## Related

[references/references-and-ownership.md](./references-and-ownership.md), [references/serialization.md](./serialization.md), [references/object-model.md](./object-model.md)
