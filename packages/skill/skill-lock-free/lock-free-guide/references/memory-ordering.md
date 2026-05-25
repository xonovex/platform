# memory-ordering: Memory Ordering

**Guideline:** Relaxed by default for independent counters; use acquire/release to publish→consume data; reserve seq*cst for code whose correctness depends on a single global order. Annotate every ordering with \_why*.

**Rationale:** The ordering argument tells the compiler and CPU how much reordering is forbidden around an atomic op. Too weak is a silent correctness bug visible only under load on weak-memory hardware (ARM/POWER); too strong is unnecessary fences and lost scaling. Choosing it is a proof obligation, not a style preference.

**The orderings:**

- `memory_order_relaxed` — atomicity and per-object modification order only; no cross-variable ordering. For counters/stats where you don't publish other data.
- `memory_order_acquire` — on a load; no reads/writes _after_ it can be hoisted before it. Pairs with a release to receive published data.
- `memory_order_release` — on a store; no reads/writes _before_ it can sink after it. Pairs with an acquire to publish data.
- `memory_order_acq_rel` — for read-modify-write (e.g. `fetch_add`, `compare_exchange`): the load half is acquire, the store half is release.
- `memory_order_seq_cst` — acquire/release _plus_ membership in a single total order over all seq_cst operations, agreed on by all threads. The default for the non-`_explicit` calls. Needed for things like Dekker/Peterson-style symmetric flag handshakes where two threads store-then-load different variables.
- `memory_order_consume` — intended as a cheaper acquire limited to data-dependent reads. Effectively deprecated: every real compiler promotes it to acquire. Do not use; write acquire.

**Acquire/release pairing (the core idiom):** a release store and an acquire load _on the same atomic_ form a synchronizes-with edge only when the acquire reads the value the release wrote. That edge makes everything sequenced-before the release visible after the acquire. One side alone guarantees nothing.

**Standalone fences:** `atomic_thread_fence(memory_order_release)` _before_ a relaxed store, and `atomic_thread_fence(memory_order_acquire)` _after_ a relaxed load, recreate the pairing without per-operation ordering. Useful to batch one fence across several relaxed ops in a hot loop. `atomic_thread_fence(memory_order_seq_cst)` is a full barrier.

**Example:**

```c
atomic_int hits = 0;
atomic_int spinlock = 0;

void record_hit(void) {
    // relaxed: a statistics counter publishes no other memory;
    // we only need the increment to be atomic and eventually counted.
    atomic_fetch_add_explicit(&hits, 1, memory_order_relaxed);
}

void unlock(void) {
    // release: everything done in the critical section must be visible
    // to the next thread that acquires this lock.
    atomic_store_explicit(&spinlock, 0, memory_order_release);
}
```

**Pitfalls:**

- An acquire that pairs with the _wrong_ variable's release synchronizes nothing.
- Mixing relaxed on one side and acquire on the other does not form a pair — both ends must be at least release/acquire.
- seq*cst is \_not* a cure-all: it orders only seq_cst ops with each other, and it's the easiest place to silently kill scalability. Start with it to be safe, then weaken with a written argument.

**Related:** [references/memory-model.md](./memory-model.md), [references/atomics-and-cas.md](./atomics-and-cas.md)
