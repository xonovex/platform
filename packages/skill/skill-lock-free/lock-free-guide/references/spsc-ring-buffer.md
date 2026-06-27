# spsc-ring-buffer: SPSC Ring Buffer

## Guideline

For exactly one producer thread and one consumer thread, use a bounded ring buffer with plain atomic head/tail indices and acquire/release ordering — no CAS, no locks.

## Rationale

With a single writer per index there is never a race _on that index_, so no read-modify-write is needed: the producer owns `tail`, the consumer owns `head`. The only cross-thread requirement is visibility — the consumer must see the slot's data before it sees the advanced index, and vice versa — which acquire/release provides exactly. This is the cheapest possible queue and the backbone of audio rings and per-worker job mailboxes.

## How to Apply

1. Size the buffer to a power of two so index wrap is a mask, not a modulo.
2. Producer: read its own `tail` relaxed, read the other side's `head` acquire to check fullness, write the slot, then publish with a release store to `tail`.
3. Consumer: mirror it — read own `head` relaxed, read `tail` acquire to check emptiness, read the slot, release-store `head`.
4. Put `head` and `tail` on separate cache lines (see false-sharing).

## Example

```c
#include <stdatomic.h>
#include <stdbool.h>
#include <stddef.h>

#define CACHE_LINE_SIZE 64        // platform cache line: 64 B x86-64, 128 B Apple silicon
#define RING_CAP 1024u            // power of two
#define RING_MASK (RING_CAP - 1u)

typedef struct {
    _Alignas(CACHE_LINE_SIZE) atomic_size_t head;  // consumer writes
    _Alignas(CACHE_LINE_SIZE) atomic_size_t tail;  // producer writes
    void *slots[RING_CAP];
} spsc_ring_t;

// Producer thread only.
bool spsc_push(spsc_ring_t *r, void *item) {
    size_t t = atomic_load_explicit(&r->tail, memory_order_relaxed); // we own tail
    size_t h = atomic_load_explicit(&r->head, memory_order_acquire); // see consumer's frees
    if (t - h >= RING_CAP) return false;          // full (unsigned wrap-safe)
    r->slots[t & RING_MASK] = item;               // write slot first...
    atomic_store_explicit(&r->tail, t + 1, memory_order_release); // ...then publish
    return true;
}

// Consumer thread only.
bool spsc_pop(spsc_ring_t *r, void **out) {
    size_t h = atomic_load_explicit(&r->head, memory_order_relaxed); // we own head
    size_t t = atomic_load_explicit(&r->tail, memory_order_acquire); // see producer's writes
    if (h == t) return false;                     // empty
    *out = r->slots[h & RING_MASK];               // read slot, ordered after acquire
    atomic_store_explicit(&r->head, h + 1, memory_order_release);    // free the slot
    return true;
}
```

## Gotchas

- Use free-running indices (never reset to 0) and unsigned subtraction for fullness/emptiness; this is wrap-safe and avoids the "is it full or empty?" ambiguity of equal head==tail.
- The `release` on `tail` pairs with the consumer's `acquire` on `tail`: that is what makes the slot write visible before the item is consumed. Swap either to relaxed and you have a data race on `slots`.
- Strictly one producer and one consumer. Two producers racing on `tail` corrupt it — that needs the MPSC or MPMC design.

## Related

[references/false-sharing.md](./false-sharing.md), [references/mpsc-queue.md](./mpsc-queue.md), [references/mpmc-bounded-queue.md](./mpmc-bounded-queue.md)
