# memory-ordering: Memory Ordering

## Guideline

Pick the weakest ordering that is correct; reserve seq_cst for code whose correctness depends on a single global total order over all seq_cst operations. Annotate every ordering with _why_.

## Rationale

The ordering argument tells the compiler and CPU how much reordering is forbidden around an atomic op. Too weak is a silent correctness bug visible only under load on weak-memory hardware (ARM/POWER); too strong is unnecessary fences and lost scaling. Choosing it is a proof obligation, not a style preference.

## The orderings

- `memory_order_relaxed` — atomicity and per-object modification order only; no cross-variable ordering.
- `memory_order_acquire` — on a load; no reads/writes _after_ it can be hoisted before it. Pairs with a release to receive published data.
- `memory_order_release` — on a store; no reads/writes _before_ it can sink after it. Pairs with an acquire to publish data.
- `memory_order_acq_rel` — for read-modify-write (e.g. `fetch_add`, `compare_exchange`): the load half is acquire, the store half is release.
- `memory_order_seq_cst` — acquire/release _plus_ membership in a single total order over all seq_cst operations, agreed on by all threads. The default for the non-`_explicit` calls. Needed for things like Dekker/Peterson-style symmetric flag handshakes where two threads store-then-load different variables.
- `memory_order_consume` — intended as a cheaper acquire limited to data-dependent reads. Effectively deprecated: every real compiler promotes it to acquire. Do not use; write acquire.

## Standalone fences

`atomic_thread_fence(memory_order_release)` _before_ a relaxed store, and `atomic_thread_fence(memory_order_acquire)` _after_ a relaxed load, recreate acquire/release pairing without per-operation ordering. Useful to batch one fence across several relaxed ops in a hot loop. `atomic_thread_fence(memory_order_seq_cst)` is a full barrier.

## Related

[references/memory-model.md](./memory-model.md), [references/atomics-and-cas.md](./atomics-and-cas.md)
