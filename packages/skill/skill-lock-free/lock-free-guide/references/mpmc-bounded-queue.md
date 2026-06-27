# mpmc-bounded-queue: Bounded MPMC Queue

## Guideline

For the general multi-producer/multi-consumer case, use a bounded queue: a power-of-two array of cells, each carrying its own sequence number, with a CAS on the position counter to claim a slot. It is fast, ABA-free, and needs no node allocation or reclamation.

## Rationale

A naive MPMC queue built from a linked list pays for allocation, CAS contention, _and_ the ABA / reclamation problem. This design sidesteps all three: storage is a fixed ring, so nothing is ever freed; each cell's sequence number encodes "whose turn it is", so producers and consumers never collide on the same cell; and the only contended atomic is a position counter advanced by CAS. The per-cell sequence makes ABA impossible without tagging — the number only ever moves forward.

## How it works

- Cell `i` is ready to _enqueue_ when `seq == pos` (the enqueue position that targets it). After writing, the producer sets `seq = pos + 1`.
- Cell `i` is ready to _dequeue_ when `seq == pos + 1`. After reading, the consumer sets `seq = pos + buffer_size` (priming it for the next lap).
- A producer/consumer reads the cell's `seq`, compares it to its own position, and CASes the shared position forward to claim the slot. Mismatch tells it the queue is full/empty or another thread already advanced.

## Example

```c
#include <stdatomic.h>
#include <stdbool.h>
#include <stddef.h>

#define CACHE_LINE_SIZE 64 // platform cache line: 64 B x86-64, 128 B Apple silicon

typedef struct { _Atomic size_t seq; void *data; } cell_t;

typedef struct {
    cell_t *buffer;
    size_t mask;                                   // buffer_size - 1 (power of two)
    _Alignas(CACHE_LINE_SIZE) _Atomic size_t enqueue_pos;
    _Alignas(CACHE_LINE_SIZE) _Atomic size_t dequeue_pos;
} mpmc_queue_t;

void mpmc_init(mpmc_queue_t *q, cell_t *buf, size_t size /* power of two */) {
    q->buffer = buf;
    q->mask = size - 1;
    for (size_t i = 0; i < size; i++)
        atomic_store_explicit(&buf[i].seq, i, memory_order_relaxed);
    atomic_store_explicit(&q->enqueue_pos, 0, memory_order_relaxed);
    atomic_store_explicit(&q->dequeue_pos, 0, memory_order_relaxed);
}

bool mpmc_enqueue(mpmc_queue_t *q, void *item) {
    cell_t *cell;
    size_t pos = atomic_load_explicit(&q->enqueue_pos, memory_order_relaxed);
    for (;;) {
        cell = &q->buffer[pos & q->mask];
        size_t seq = atomic_load_explicit(&cell->seq, memory_order_acquire);
        intptr_t diff = (intptr_t)seq - (intptr_t)pos;
        if (diff == 0) {                            // cell is ours to fill
            // claim the slot by advancing enqueue_pos; relaxed is fine, the
            // acquire on seq above + release on seq below carry the ordering.
            if (atomic_compare_exchange_weak_explicit(
                    &q->enqueue_pos, &pos, pos + 1,
                    memory_order_relaxed, memory_order_relaxed))
                break;                              // won the race
        } else if (diff < 0) {
            return false;                           // full: consumer hasn't caught up
        } else {
            // another producer advanced enqueue_pos; reload and retry
            pos = atomic_load_explicit(&q->enqueue_pos, memory_order_relaxed);
        }
    }
    cell->data = item;                              // write payload into our claimed cell
    // release: publish payload to the consumer that will read seq == pos+1.
    atomic_store_explicit(&cell->seq, pos + 1, memory_order_release);
    return true;
}

bool mpmc_dequeue(mpmc_queue_t *q, void **out) {
    cell_t *cell;
    size_t pos = atomic_load_explicit(&q->dequeue_pos, memory_order_relaxed);
    for (;;) {
        cell = &q->buffer[pos & q->mask];
        size_t seq = atomic_load_explicit(&cell->seq, memory_order_acquire);
        intptr_t diff = (intptr_t)seq - (intptr_t)(pos + 1);
        if (diff == 0) {                            // cell holds data for us
            if (atomic_compare_exchange_weak_explicit(
                    &q->dequeue_pos, &pos, pos + 1,
                    memory_order_relaxed, memory_order_relaxed))
                break;
        } else if (diff < 0) {
            return false;                           // empty
        } else {
            pos = atomic_load_explicit(&q->dequeue_pos, memory_order_relaxed);
        }
    }
    *out = cell->data;                              // read payload, ordered by acquire
    // release: free the cell for the producer one lap ahead (pos + size).
    atomic_store_explicit(&cell->seq, pos + q->mask + 1, memory_order_release);
    return true;
}
```

## Why it's fast and ABA-free

producers and consumers contend only on their _own_ position counter (separate cache lines), never on each other's; a claimed cell is owned exclusively until its sequence is bumped, so two writers can't target it. The sequence number is strictly monotonic per cell across laps, so a CAS can never be fooled by a stale-then-recurring value — no tagging or reclamation needed.

## Gotchas

- Size must be a power of two; `mask` does the wrap. The capacity is the array size — it is genuinely bounded (returns false when full).
- `enqueue_pos` and `dequeue_pos` must live on separate cache lines or producers and consumers false-share them.
- Use `intptr_t` for the `diff` so wraparound of the unsigned positions compares correctly as a signed difference.

## Related

[references/aba-problem.md](./aba-problem.md), [references/false-sharing.md](./false-sharing.md), [references/spsc-ring-buffer.md](./spsc-ring-buffer.md)
