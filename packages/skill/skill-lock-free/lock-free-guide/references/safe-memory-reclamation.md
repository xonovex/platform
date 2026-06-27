# safe-memory-reclamation: Safe Memory Reclamation

## Guideline

In a lock-free structure, removing a node from the logical structure is not the same as freeing it; defer the actual `free` until you can prove no other thread holds a reference. Pick a reclamation scheme deliberately.

## Rationale

This is the central hard problem of lock-free programming. Without a lock, a thread can be reading node X at the exact moment another thread unlinks and frees it — a use-after-free, and the source of ABA when the address is recycled. A garbage collector solves this for you; in C you must build the equivalent. Every scheme trades reader cost, memory overhead, and reclamation latency differently.

## Schemes

- **Reference counting:** each node carries an atomic refcount; readers increment before access, decrement after, and the last to reach zero frees it. _Limit:_ the increment is itself a race — between loading the pointer and bumping its count, the node can be freed (the "load-then-inc" gap). Fixing it needs split reference counts or a DWCAS, and the per-access atomic RMW does not scale (cache-line bouncing on the count). Fine for low-contention or coarse-grained nodes.

- **Hazard pointers:** each thread publishes the pointers it is _about to_ dereference into a small per-thread array (its hazard slots) with a release store, then re-validates the pointer is still installed. A retiring thread puts the node on a private retire list; it frees a node only when scanning all threads' hazard slots shows none point at it. _Trade-off:_ O(1) bounded memory per thread, wait-free-ish reads with a store + fence per access, but a scan cost at reclaim time and a fixed cap on simultaneously-protected pointers. Good when you must bound memory.

- **Epoch-based reclamation / QSBR:** maintain a global epoch counter; a reader enters a critical section by announcing the current epoch, and retired nodes are freed only once every thread has advanced past the epoch in which they were retired (no thread could still hold them). _QSBR_ (quiescent-state-based) is the cheapest variant: readers have _zero_ per-access overhead and merely report a quiescent state periodically (e.g. once per frame). _Trade-off:_ near-free reads, but a single stalled thread that never quiesces stalls all reclamation — memory usage is unbounded until it does. Excellent for a game loop where every worker reaches a known quiescent point each frame.

- **RCU (read-copy-update):** readers run in a lightweight read-side critical section (often just disabling preemption / a per-CPU counter) and never block writers; writers publish a new version with a release store and wait for a _grace period_ — the moment all pre-existing readers have finished — before freeing the old version. Effectively epoch-based reclamation specialized for read-mostly data. _Trade-off:_ the fastest possible reads, but writers pay the grace-period latency and the pattern fits read-mostly, copy-on-write data far better than write-heavy structures.

- **Immutable snapshots (atomic version swap):** keep the shared state in an immutable object; a reader does one acquire-load of the current version pointer and then reads freely with no further synchronization (the version never mutates). A writer builds a whole new version off to the side and publishes it with one release-store/CAS on the pointer. Old versions are reclaimed (refcount or epoch) once no reader still holds them. _Trade-off:_ trivially correct lock-free reads and a simple mental model, at the cost of copying the state per update — ideal for coarse, read-mostly, infrequently-written data (config, a world/scene snapshot per frame).

## How to choose

1. Read-mostly, readers must be cheapest, can define quiescent points (frame boundaries) → QSBR / RCU.
2. Must bound worst-case memory, mixed read/write → hazard pointers.
3. Low contention, simplest to reason about, coarse objects → reference counting.
4. Fixed pool that never returns memory to the allocator → no reclamation needed; tag pointers for ABA instead.

## Gotchas

- All four protect _one logical structure_; a node shared across structures needs the strictest scheme covering it.
- QSBR/RCU correctness hinges on _every_ participating thread reaching a quiescent state; a thread that spins forever without quiescing leaks all retired memory.
- Hazard-pointer reads need a `seq_cst` fence (or store + acquire fence) between publishing the hazard and re-reading the pointer, or the publish can be reordered after the dereference.

## Related

[references/lock-free-stack.md](./lock-free-stack.md), [references/aba-problem.md](./aba-problem.md), [references/progress-guarantees.md](./progress-guarantees.md)
