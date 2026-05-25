# locks-and-backoff: Locks and Backoff

**Guideline:** A lock is often the right answer; when you spin, spin politely (test-and-test-and-set + exponential backoff + `pause`), and when contention is high reach for a queue lock that doesn't bounce a single cache line across cores.

**Rationale:** Lock-free is not automatically better — an uncontended or well-designed lock is frequently faster and always simpler. The performance failures of locks come from _how_ threads wait: a naive spinlock that hammers an atomic exchange in a tight loop saturates the coherence bus and starves the holder. The fixes are well known; pick by contention level and wait duration.

**Spinlock — test-and-test-and-set with backoff:**

```c
#include <stdatomic.h>
#include <stdbool.h>

atomic_bool locked;  // false = unlocked. Use atomic_bool, not atomic_flag,
                     // so we can spin on a cheap load (atomic_flag has no load).

static inline void cpu_relax(void) {
#if defined(__x86_64__) || defined(__i386__)
    __builtin_ia32_pause();   // x86 PAUSE: hint, saves power, frees the pipeline
#elif defined(__aarch64__)
    __asm__ __volatile__("yield");
#endif
}

void spin_lock(atomic_bool *l) {
    unsigned backoff = 1;
    for (;;) {
        // Test: spin on a relaxed *load* (stays in shared cache state, no
        // coherence traffic) until the lock looks free.
        while (atomic_load_explicit(l, memory_order_relaxed)) {
            for (unsigned i = 0; i < backoff; i++) cpu_relax();
            if (backoff < 1024) backoff <<= 1;       // exponential backoff
        }
        // ...then-set: one acquire exchange to actually claim it.
        if (!atomic_exchange_explicit(l, true, memory_order_acquire))
            return;                                  // we observed false, now own it
        backoff = 1;                                 // lost the race; reset and retry
    }
}

void spin_unlock(atomic_bool *l) {
    atomic_store_explicit(l, false, memory_order_release); // publish the section
}
```

**Lock varieties — pick by contention and pattern:**

- **TTAS spinlock + backoff:** short critical sections, low/medium contention. Simple, but all waiters still contend on one line.
- **Ticket lock:** FIFO fairness — take a `fetch_add` ticket, wait until `now_serving` equals it. Fair, but every waiter spins on the same `now_serving` line (cache-line bouncing under high contention).
- **MCS lock:** each waiter spins on its _own_ queue node's flag; the unlocker hands off to the next node. No global bouncing, scales to many cores; needs a per-acquire node (often stack-local).
- **CLH lock:** like MCS but each waiter spins on its _predecessor's_ node; slightly simpler, implicit queue. Both MCS and CLH are the scalable choice for contended locks.
- **Seqlock:** read-mostly data. Writer bumps an even→odd→even counter around its update; readers snapshot the counter, read, and retry if it changed or was odd. Lock-free, wait-free reads with no writer-visible cost — but readers must tolerate reading torn data and retrying, so the protected data must be trivially copyable.
- **rwlock:** many readers / occasional writer when readers must _block_ writers (can't just retry). Heavier than a seqlock; prefer seqlock or RCU when readers dominate.
- **Eventcount:** lets threads _block_ (futex/condvar) when there's no work instead of burning CPU spinning, while keeping a lock-free fast path. The right pattern for an idle worker pool: spin briefly, then sleep on the eventcount until producers signal.

**Gotchas:**

- A spinlock is correct only if the holder cannot be preempted for long; on a loaded machine prefer a blocking mutex or an eventcount so a descheduled holder doesn't make others burn cycles.
- `pause`/`yield` are not optional micro-optimizations on hyperthreaded cores — without them spinning starves the sibling thread (which may be the holder).
- Always `release` on unlock and `acquire` on lock; that pairing is what publishes the critical section's writes to the next holder.

**Related:** [references/progress-guarantees.md](./progress-guarantees.md), [references/scalable-architecture.md](./scalable-architecture.md), [references/memory-ordering.md](./memory-ordering.md)
