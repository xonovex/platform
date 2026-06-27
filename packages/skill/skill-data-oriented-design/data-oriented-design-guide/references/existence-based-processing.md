# existence-based-processing: Existence-Based Processing

## Guideline

Process every item that exists together by grouping items by type or state, so per-item conditionals disappear and branches become loops.

## Rationale

The principle "the absence of a thing is a thing": instead of every item carrying a flag the hot loop must branch on, store items in collections defined by their state — membership in a collection _is_ the flag. This removes the per-item branch (no misprediction stalls, no wasted work skipping dead items) and turns a heterogeneous, branch-laden loop into several homogeneous, branch-free loops. The loop also only touches live data, shrinking the working set.

## Techniques

- **Bucket by type/state** - Keep a separate array per type or state (active/sleeping, alive/dead, visible/culled). Iterate the whole bucket with no `if`.
- **Sort/partition to make branches into loops** - Partition a mixed array so all items in state A precede all in state B; then process each contiguous run with one specialized loop.
- **Boolean / bitset arrays** - Represent "does this slot have component X" as a packed bit array; scan it to build the index list of present items, then stream those.
- **Absence is a thing** - Rather than nulling a field and checking it, move the item out of the "has it" collection entirely. The collection's membership answers the question.
- **Move on transition** - When an item changes state, move it between buckets at that moment (rare event) instead of testing its state every frame (common event).

## How to Apply

1. Find a hot loop with a per-item `if (item.state == X)` / `if (item.alive)`.
2. Replace the flag with collection membership: maintain one bucket per state.
3. On state change, swap-remove the item from its old bucket and append to the new one.
4. Rewrite the loop as a branch-free sweep over the relevant bucket(s).

## Example

```c
// Bad: branch per item every frame; mispredicts; touches dead entries.
for (size_t i = 0; i < n; i++)
  if (e[i].alive && e[i].state == STATE_MOVING)
    e[i].pos = step(e[i].pos);

// Good: membership encodes state; loop is branch-free and touches only movers.
// moving[] holds indices/records of exactly the moving entities.
for (size_t i = 0; i < moving_count; i++)
  moving[i].pos = step(moving[i].pos);

// transition (rare): item stops moving -> move it out of the moving bucket.
static void stop(world_t *w, uint32_t i) {
  w->moving[i] = w->moving[--w->moving_count]; // swap-remove
  w->idle[w->idle_count++] = /* the stopped entity */;
}
```

## Gotchas

- The cost moves to the transition: bucketing pays off when items change state far less often than they are processed.
- Many buckets with few items each can fragment the working set; balance bucket count against batch size.
- Keep buckets consistent with any handle/index references (swap-remove changes positions).

## Related

[references/data-as-transforms.md](./data-as-transforms.md), [references/handles-and-indices.md](./handles-and-indices.md), [references/access-patterns.md](./access-patterns.md), [references/soa-aos-aosoa.md](./soa-aos-aosoa.md)
