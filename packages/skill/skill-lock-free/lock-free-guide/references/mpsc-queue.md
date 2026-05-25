# mpsc-queue: Multi-Producer Single-Consumer Queue

**Guideline:** For many producers feeding one consumer, use an intrusive MPSC queue: producers publish with a single atomic exchange on the tail; the consumer walks `next` pointers — no CAS, no per-node allocation in the queue itself.

**Rationale:** The expensive part of a concurrent queue is the producer-side contention. Replacing a CAS loop with one unconditional `atomic_exchange` makes each enqueue _wait-free_ (one bounded step, never retries) while keeping the consumer trivially single-threaded. It is the natural fit for log sinks, job-completion collection, and event funnels where N workers report to one drainer.

**How it works:**

1. Nodes are intrusive (caller embeds a `next` field) and the queue is initialized with a dummy "stub" node so it is never empty internally.
2. `enqueue`: set the node's `next` to NULL, atomically swap it into `tail` (release), then link the _previous_ tail's `next` to the new node. Between the swap and the link the queue is momentarily inconsistent — the chain has a transient gap.
3. `dequeue` (consumer only): follow `head->next`; if it's NULL the queue may be empty _or_ a producer is mid-enqueue (the gap) — report empty and retry later.

**Example:**

```c
#include <stdatomic.h>
#include <stddef.h>

#define CACHE_LINE_SIZE 64 // platform cache line: 64 B x86-64, 128 B Apple silicon

typedef struct mpsc_node { _Atomic(struct mpsc_node *) next; } mpsc_node_t;

typedef struct {
    _Alignas(CACHE_LINE_SIZE) _Atomic(mpsc_node_t *) tail; // producers
    _Alignas(CACHE_LINE_SIZE) mpsc_node_t *head;           // consumer only
    mpsc_node_t stub;
} mpsc_queue_t;

void mpsc_init(mpsc_queue_t *q) {
    atomic_store_explicit(&q->stub.next, NULL, memory_order_relaxed);
    q->head = &q->stub;
    atomic_store_explicit(&q->tail, &q->stub, memory_order_relaxed);
}

// Any producer thread. Wait-free: one exchange, no loop.
void mpsc_enqueue(mpsc_queue_t *q, mpsc_node_t *n) {
    atomic_store_explicit(&n->next, NULL, memory_order_relaxed);
    // release: the node's contents must be visible to the consumer that
    // later acquires this link. exchange returns the previous tail.
    mpsc_node_t *prev = atomic_exchange_explicit(&q->tail, n, memory_order_acq_rel);
    // GAP: until this store, prev->next is NULL and the chain is broken.
    atomic_store_explicit(&prev->next, n, memory_order_release);
}

// Consumer thread only. Returns NULL if empty or a producer is mid-enqueue.
mpsc_node_t *mpsc_dequeue(mpsc_queue_t *q) {
    mpsc_node_t *head = q->head;
    mpsc_node_t *next = atomic_load_explicit(&head->next, memory_order_acquire);
    if (next == NULL) return NULL;     // empty, or transient gap: try again later
    q->head = next;                    // advance; `head` (old stub) is now free to reuse
    return head;                       // payload is in the node that `next` displaced
}
```

**Gotchas:**

- The transient gap means `dequeue` can spuriously return NULL while the queue is non-empty; treat NULL as "nothing _right now_", never as a permanent empty signal.
- It is wait-free for _producers_ but only lock-free for the _consumer_: a producer descheduled inside the gap stalls the consumer until it resumes. Acceptable for a single trusted consumer; never expose two consumers.
- The returned node is the _old_ head/stub; the freshly enqueued payload lives in the node it pointed to. Many implementations copy the payload out of `next` and recycle the old node as the new stub — be consistent about which node carries the data.

**Related:** [references/spsc-ring-buffer.md](./spsc-ring-buffer.md), [references/atomics-and-cas.md](./atomics-and-cas.md), [references/mpmc-bounded-queue.md](./mpmc-bounded-queue.md)
