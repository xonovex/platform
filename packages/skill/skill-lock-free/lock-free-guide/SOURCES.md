# Sources

## Dmitry Vyukov — 1024cores

- **URL:** https://www.1024cores.net/ (canonical home of the lock-free/wait-free articles; the older `sites.google.com/site/1024cores/home` mirror is superseded)
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Building blocks, Data structures, Reclamation & scaling
  - The queue/stack algorithms and the Relacy verification guidance under `references/`
- **Aspects extracted:**
  - Lock-free vs wait-free vs obstruction-free taxonomy → `references/progress-guarantees.md`
  - MPSC intrusive queue → `references/mpsc-queue.md`
  - Bounded MPMC queue (per-cell sequence numbers) → `references/mpmc-bounded-queue.md`
  - Cache-line / false-sharing guidance → `references/false-sharing.md`
  - Relacy Race Detector → `references/testing-and-verification.md`
  - Scalability / "the fastest synchronization is none" → `references/scalable-architecture.md`

## C11 / C++11 Memory Model

- **URL:** https://en.cppreference.com/w/c/atomic and https://en.cppreference.com/w/c/atomic/memory_order
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Memory model section
  - All `<stdatomic.h>` usage in `references/`
- **Aspects extracted:**
  - Data race = UB, happens-before, synchronizes-with, modification order → `references/memory-model.md`
  - `memory_order_*` semantics and fences → `references/memory-ordering.md`
  - Atomic load/store/RMW, `compare_exchange_weak`/`_strong`, spurious failure → `references/atomics-and-cas.md`

## Maurice Herlihy & Nir Shavit — The Art of Multiprocessor Programming

- **URL:** https://www.elsevier.com/books/the-art-of-multiprocessor-programming/herlihy/978-0-12-415950-1 (2nd ed.)
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Memory model, Data structures, Reclamation & scaling
- **Aspects extracted:**
  - Progress conditions, wait-free/lock-free hierarchy → `references/progress-guarantees.md`
  - Treiber stack and the ABA problem → `references/lock-free-stack.md`, `references/aba-problem.md`
  - Spinlocks, TTAS, exponential backoff, MCS/CLH queue locks → `references/locks-and-backoff.md`

## Paul E. McKenney — Is Parallel Programming Hard, And, If So, What Can You Do About It?

- **URL:** https://mirrors.edge.kernel.org/pub/linux/kernel/people/paulmck/perfbook/perfbook.html
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Reclamation & scaling
- **Aspects extracted:**
  - RCU, QSBR, hazard pointers, reference counting trade-offs → `references/safe-memory-reclamation.md`
  - Seqlock, per-CPU/per-thread sharding, partitioning, NUMA → `references/scalable-architecture.md`, `references/locks-and-backoff.md`

## Game-engine development blog (archive)

- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Reclamation & scaling
  - Practical engine concurrency patterns
- **Aspects extracted:**
  - "Fiber based job system" — fibers yield instead of blocking a worker → `references/scalable-architecture.md`
  - "Multi-Threading The Truth" — immutable snapshots / atomic version swap for read-mostly data → `references/safe-memory-reclamation.md`

## Refresh Workflow

1. Re-fetch the upstream source(s)
2. Diff against the prior pull (or scan for newly added sections)
3. For each changed area, update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
