# memory-model: The C11 Memory Model

**Guideline:** A data race is undefined behavior; make every location that two threads access (where at least one writes) either atomic or protected by a lock that establishes happens-before.

**Rationale:** The compiler and CPU freely reorder, fuse, and cache non-atomic memory operations on the assumption that no other thread observes them. If two threads touch the same non-atomic object concurrently and at least one writes, the program has a data race and the entire execution is UB — not "a stale value", but no defined behavior at all. Atomics and locks exist to create the ordering relations that make one thread's writes _defined to be visible_ to another.

**Key relations:**

- **Sequenced-before:** ordering within a single thread (roughly, program order between evaluations).
- **Synchronizes-with:** the cross-thread edge. A release operation on an atomic _synchronizes-with_ an acquire operation on the same atomic that reads the released value (or a value later in its modification order, via a release sequence).
- **Happens-before:** the transitive closure of sequenced-before and synchronizes-with. If write A happens-before read B, then B is guaranteed to see A (and everything sequenced-before A). No happens-before edge → no visibility guarantee.
- **Modification order:** for each atomic object, all writes to it form a single total order that every thread agrees on, even under relaxed ordering. Relaxed gives you this per-object order — nothing more.

**Atomic vs non-atomic:** only `_Atomic`-qualified objects (or those accessed via `atomic_*`) may be touched concurrently. A non-atomic object published _through_ a release/acquire pair is fine — the synchronizes-with edge orders the non-atomic write before the reader's non-atomic read.

**Example:**

```c
#include <stdatomic.h>

int payload;                       // plain, non-atomic
atomic_bool ready = false;

// Producer
void produce(void) {
    payload = 42;                            // non-atomic write...
    atomic_store_explicit(&ready, true,
                          memory_order_release); // ...published here
}

// Consumer
int consume(void) {
    // acquire pairs with the release: if we observe ready == true,
    // the producer's payload write happens-before this point.
    while (!atomic_load_explicit(&ready, memory_order_acquire))
        ; // spin
    return payload;                          // guaranteed to read 42
}
```

**Why atomics, not `volatile`:** `volatile` only stops the _compiler_ from eliding/reordering accesses to that object. It establishes no happens-before, emits no CPU barriers, and is not even guaranteed indivisible. It is for memory-mapped I/O and signal handlers, never for inter-thread synchronization. Use `_Atomic` / `atomic_*`.

**Related:** [references/memory-ordering.md](./memory-ordering.md), [references/atomics-and-cas.md](./atomics-and-cas.md)
