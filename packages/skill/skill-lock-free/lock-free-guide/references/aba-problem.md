# aba-problem: The ABA Problem

**Guideline:** Never assume a successful CAS means the location was untouched; if a value can change from A to B and back to A while a thread sleeps, pair the pointer with a version counter (or eliminate reuse with a reclamation scheme).

**Rationale:** CAS compares bit patterns. If thread T reads head = A, stalls, and meanwhile another thread pops A, pushes B, and pushes A again (often the _same freed-then-reallocated node_), T's CAS still sees A and succeeds — but the surrounding structure has silently changed. The classic failure is a lock-free stack pop that swings `head` to a stale `A->next`, corrupting the list or touching freed memory.

**Why CAS alone is insufficient:** the danger appears specifically when memory is _reclaimed and reused_. Without reclamation a node is never freed, so address A can't be recycled into a different logical node mid-operation; with reclamation, A's address can mean different things at different times, and CAS cannot tell them apart.

**Mitigations:**

- **Tagged / versioned pointers (DWCAS):** store the pointer next to a monotonically increasing counter and CAS both atomically. Any A→B→A cycle bumps the tag, so the stale CAS fails. Needs a double-width CAS — on x86-64 that is `CMPXCHG16B` (128-bit); C11 expresses it as an atomic 16-byte struct.

```c
#include <stdatomic.h>
typedef struct { void *ptr; uintptr_t tag; } tagged_ptr_t;
// 16-byte struct; verify lock-freedom on the target ABI.
_Atomic tagged_ptr_t head;

bool push_tagged(node_t *n) {
    tagged_ptr_t cur = atomic_load_explicit(&head, memory_order_acquire);
    tagged_ptr_t neu;
    do {
        n->next = cur.ptr;
        neu.ptr = n;
        neu.tag = cur.tag + 1;   // bump version: defeats any A->B->A
    } while (!atomic_compare_exchange_weak_explicit(
                 &head, &cur, neu,
                 memory_order_release, memory_order_acquire));
    return true;
}
```

- **Pack the tag into spare pointer bits** when DWCAS is unavailable: aligned pointers leave low bits free; a few bits of tag shrink the cycle window but don't fully close it.
- **Avoid reuse via a reclamation scheme:** hazard pointers, epoch/QSBR, or RCU guarantee a node is not freed (hence not recycled) while any thread might still CAS against it — removing the ABA precondition entirely. This is the general fix for arbitrary node graphs.

**Gotchas:**

- A tag counter can in theory wrap; make it wide (≥ pointer-sized) so wraparound is unreachable in practice.
- `_Atomic` on a 16-byte struct is only lock-free if the target supports DWCAS — check `atomic_is_lock_free`, or it silently falls back to a hidden mutex.
- Tagging fixes ABA but _not_ use-after-free: a versioned CAS can still dereference `cur.ptr` after the node was freed. ABA-freedom and reclamation are separate problems.

**Related:** [references/lock-free-stack.md](./lock-free-stack.md), [references/safe-memory-reclamation.md](./safe-memory-reclamation.md), [references/mpmc-bounded-queue.md](./mpmc-bounded-queue.md)
