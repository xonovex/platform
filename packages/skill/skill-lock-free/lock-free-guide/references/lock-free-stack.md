# lock-free-stack: Lock-Free Stack

## Guideline

Build a lock-free LIFO by CAS-ing a new node onto `head`; but recognize that the _pop_ path has both an ABA hazard and a use-after-free hazard, so a lock-free stack that frees nodes REQUIRES a memory-reclamation scheme.

## Rationale

This is the canonical lock-free structure: push and pop are each a short CAS loop on a single `head` pointer. Push is safe in isolation. Pop, however, dereferences `head->next` to compute the new head — and between that read and the CAS another thread can pop and free the very node you're holding, giving (a) a torn/incorrect `next` (ABA) and (b) a dereference of freed memory. This is the textbook example of why "lock-free" and "safe to free" are different problems.

## Example

```c
#include <stdatomic.h>
#include <stdbool.h>

typedef struct node { struct node *next; void *value; } node_t;
_Atomic(node_t *) head;

// Push is safe on its own: we only ever read head, never *head's successor*.
void push(node_t *n) {
    node_t *old = atomic_load_explicit(&head, memory_order_relaxed);
    do {
        n->next = old;
        // release: publish n->value/n->next before n becomes reachable.
    } while (!atomic_compare_exchange_weak_explicit(
                 &head, &old, n, memory_order_release, memory_order_relaxed));
}

// Pop is UNSAFE as written if popped nodes can be freed by other threads.
node_t *pop_UNSAFE(void) {
    node_t *old = atomic_load_explicit(&head, memory_order_acquire);
    while (old != NULL) {
        // HAZARD: another thread may pop & free `old` here, so this read of
        // old->next is a use-after-free, and a freed-then-reallocated `old`
        // makes the CAS below succeed against a stale list (ABA).
        node_t *next = old->next;
        if (atomic_compare_exchange_weak_explicit(
                &head, &old, next, memory_order_acquire, memory_order_acquire))
            return old;     // caller must NOT free immediately if others may hold it
    }
    return NULL;
}
```

## How to Apply (making pop safe)

1. Protect `old` before dereferencing it — publish it as a _hazard pointer_ so no thread frees it while you hold it, then re-validate `head` still equals it.
2. Or wrap pop in an _epoch / RCU read-side section_ and defer frees until all readers have left the epoch.
3. Or, if the stack is a fixed pool that never frees (a freelist of preallocated nodes), tag `head` with a version counter (DWCAS) to kill ABA — nodes are reused but never returned to the allocator, so there is no use-after-free.

## Gotchas

- Tagging alone (versioned head) fixes ABA but _not_ use-after-free; it is only sufficient when nodes are never actually freed.
- Don't `free(old)` right after a successful pop in a shared stack — another thread may still be inside its own pop holding `old`. Hand it to the reclamation scheme.
- Push being safe is specific to LIFO: the node is fully initialized before the CAS and no concurrent reader walks past `head` during push.

## Related

[references/safe-memory-reclamation.md](./safe-memory-reclamation.md), [references/aba-problem.md](./aba-problem.md), [references/atomics-and-cas.md](./atomics-and-cas.md)
