# progress-guarantees: Progress Guarantees

**Guideline:** Classify every concurrent algorithm by its progress guarantee, and only pay for a stronger one when the workload demands it.

**Rationale:** "Lock-free" is a precise liveness property, not a synonym for "fast" or "no mutex". Blocking code can stall the whole system if one thread is descheduled while holding a lock; non-blocking code cannot. But stronger guarantees usually cost more per operation, so the right choice depends on whether you fear latency tails, priority inversion, or just want throughput.

**Techniques:**

- **Blocking:** progress requires another thread to act (release a lock). A preempted, crashed, or paged-out lock holder halts everyone — deadlock, livelock, priority inversion, convoying all live here.
- **Obstruction-free:** a thread makes progress if it runs in isolation (no contention). Weakest non-blocking class; can livelock under contention without a helping/backoff scheme.
- **Lock-free (system-wide progress):** out of all contending threads, _at least one_ always completes in a bounded number of steps. Individual threads may starve (their CAS keeps losing), but the system never stalls. No lock is held, so a stalled thread never blocks others.
- **Wait-free (per-thread progress):** _every_ thread completes in a bounded number of its own steps, regardless of others. Strongest guarantee — no starvation, bounded worst-case latency. Usually needs helping schemes or fetch-add-style primitives; often slower in the common case.
- **Containment:** wait-free ⊃ lock-free ⊃ obstruction-free ⊃ blocking. A wait-free algorithm is also lock-free, etc.

**When each matters:**

- Hard real-time / audio / no-malloc render thread → wait-free or at least lock-free (a frame must never stall on another thread's lock).
- General throughput / server work → a good lock or lock-free; lock-free helps mainly under high contention or when holders can be preempted.
- Read-mostly shared config → seqlock or RCU (lock-free reads) beats a rwlock.

**Trade-offs vs locks:**

- Locks are simpler, compose, and are frequently _faster_ uncontended; a well-tuned lock beats a sloppy lock-free queue.
- Lock-free removes lock-holder-stall failure modes (priority inversion, convoying) and degrades more gracefully under preemption.
- Lock-free does not remove the hard problems: ABA and safe memory reclamation move the complexity, they don't delete it. Reach for lock-free only with a measured reason.

**Related:** [references/locks-and-backoff.md](./locks-and-backoff.md), [references/safe-memory-reclamation.md](./safe-memory-reclamation.md), [references/scalable-architecture.md](./scalable-architecture.md)
