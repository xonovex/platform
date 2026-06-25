---
name: lock-free-guide
description: "Use when writing or reviewing shared-memory concurrent code: atomics, lock-free/wait-free data structures, or scalable synchronization. Triggers on prompts about memory ordering (relaxed/acquire/release/seq_cst), CAS/compare-exchange, the ABA problem, SPSC/MPSC/MPMC queues, lock-free stack, hazard pointers / epoch / RCU reclamation, false sharing, spinlocks/backoff, work-stealing, even when the user doesn't say 'lock-free'."
---

# Lock-Free / Wait-Free Concurrency Guidelines

## Requirements

- C11 atomics (`<stdatomic.h>`); audience is a C99/C11 game engine with atomic/thread/job modules.

## Essentials

- **Default relaxed, escalate deliberately** - Pick the weakest ordering that is correct, see [references/memory-ordering.md](references/memory-ordering.md)
- **No data races** - Every shared mutable location touched by 2+ threads is atomic or lock-protected, see [references/memory-model.md](references/memory-model.md)
- **CAS loops retry** - `compare_exchange_weak` in a loop; reload the expected value on failure, see [references/atomics-and-cas.md](references/atomics-and-cas.md)
- **Lock-free needs reclamation** - Removing a node from a shared structure is not freeing it, see [references/safe-memory-reclamation.md](references/safe-memory-reclamation.md)
- **Measure, do not assume** - Locks are often faster; profile before going lock-free, see [references/progress-guarantees.md](references/progress-guarantees.md)

## Memory model

- **Progress guarantees** - wait-free ⊃ lock-free ⊃ obstruction-free ⊃ blocking, see [references/progress-guarantees.md](references/progress-guarantees.md)
- **Happens-before** - Acquire/release builds the ordering that makes writes visible, see [references/memory-model.md](references/memory-model.md)
- **Orderings** - relaxed/acquire/release/acq_rel/seq_cst + fences, see [references/memory-ordering.md](references/memory-ordering.md)

## Building blocks

- **Atomics & CAS** - load/store/exchange/fetch\_\*/compare_exchange, weak vs strong, see [references/atomics-and-cas.md](references/atomics-and-cas.md)
- **ABA problem** - CAS can't see a value changed away and back; tag pointers or reclaim safely, see [references/aba-problem.md](references/aba-problem.md)
- **False sharing** - Pad/align hot per-thread and atomic fields to their own cache line, see [references/false-sharing.md](references/false-sharing.md)

## Data structures

- **SPSC ring** - One producer, one consumer, no CAS, see [references/spsc-ring-buffer.md](references/spsc-ring-buffer.md)
- **MPSC queue** - Intrusive: producers XCHG tail, consumer walks next, see [references/mpsc-queue.md](references/mpsc-queue.md)
- **MPMC bounded** - Per-cell sequence numbers, ABA-free, see [references/mpmc-bounded-queue.md](references/mpmc-bounded-queue.md)
- **Lock-free stack** - CAS on head; needs reclamation for pop, see [references/lock-free-stack.md](references/lock-free-stack.md)

## Reclamation & scaling

- **Safe reclamation** - refcount / hazard pointers / epoch (QSBR) / RCU, see [references/safe-memory-reclamation.md](references/safe-memory-reclamation.md)
- **Locks & backoff** - TTAS spinlock, exponential backoff, ticket/MCS/CLH, seqlock, see [references/locks-and-backoff.md](references/locks-and-backoff.md)
- **Scale by not sharing** - Shard, partition, combine, work-steal, NUMA-place, see [references/scalable-architecture.md](references/scalable-architecture.md)

## Gotchas

- A successful CAS proves only that the bit pattern matched — not that nothing happened in between (ABA); pointers reused after free defeat it.
- `volatile` is not atomic and gives no ordering — it never makes concurrent C correct; use `_Atomic`/`atomic_*`.
- `compare_exchange_weak` may fail spuriously even when the value matches — only ever use it inside a retry loop.
- `seq_cst` is the safe default for a first draft but the slowest; relaxing it later requires a proof, not a guess.
- Removing a node from a lock-free list does not mean no other thread is still dereferencing it — freeing too early is a use-after-free.
- Two unrelated atomics on the same cache line still contend (false sharing) — alignment, not correctness, but it can erase all your scaling.

## Progressive Disclosure

- Read [references/progress-guarantees.md](references/progress-guarantees.md) - Load when deciding between locks, lock-free, or wait-free, or explaining the trade-offs
- Read [references/memory-model.md](references/memory-model.md) - Load when reasoning about data races, visibility, or happens-before
- Read [references/memory-ordering.md](references/memory-ordering.md) - Load when choosing or reviewing a `memory_order_*` argument or a fence
- Read [references/atomics-and-cas.md](references/atomics-and-cas.md) - Load when writing atomic loads/stores/RMW or a compare-exchange loop
- Read [references/aba-problem.md](references/aba-problem.md) - Load when a CAS-based structure reuses memory or pointers
- Read [references/false-sharing.md](references/false-sharing.md) - Load when per-thread counters or atomics scale badly under contention
- Read [references/spsc-ring-buffer.md](references/spsc-ring-buffer.md) - Load when building a single-producer/single-consumer queue
- Read [references/mpsc-queue.md](references/mpsc-queue.md) - Load when many threads feed one consumer (logging, job results)
- Read [references/mpmc-bounded-queue.md](references/mpmc-bounded-queue.md) - Load when building a general bounded multi-producer/multi-consumer queue
- Read [references/lock-free-stack.md](references/lock-free-stack.md) - Load when implementing a lock-free stack or freelist
- Read [references/safe-memory-reclamation.md](references/safe-memory-reclamation.md) - Load when freeing nodes that other threads might still read
- Read [references/locks-and-backoff.md](references/locks-and-backoff.md) - Load when a lock is the right tool and you need a scalable one
- Read [references/scalable-architecture.md](references/scalable-architecture.md) - Load when synchronization itself is the bottleneck
- Read [references/testing-and-verification.md](references/testing-and-verification.md) - Load when validating concurrent code before merge
