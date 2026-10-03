# handles-and-indices: Handles and Indices Instead of Pointers

## Guideline

Reference entities by integer index or generational handle (index + generation) into a contiguous array, never by raw pointer.

## Rationale

Raw pointers pin objects to fixed addresses, so the backing storage cannot be relocated, compacted, or serialized, and a freed object leaves dangling pointers. Indices are small, relocatable (the array can move or grow), serialization-stable, and keep the data in a cache-friendly contiguous block.

## Techniques

- **Index reference** - Store a `uint32_t` index into the array instead of a pointer. Half the size of a 64-bit pointer; doubles references per cache line.
- **Swap-remove** - To delete element `i` in O(1) without holes, move the last element into slot `i` and shrink count. Keeps the array dense and contiguous for streaming.
- **Free list** - To keep stable slot indices, reuse freed slots via an intrusive free list (store the next-free index in the dead slot); bump generation on reuse.
- **Stable vs dense** - Swap-remove gives a dense array but moving indices (fix up references or use handles); free-list gives stable indices but a possibly sparse array (use an alive flag, see existence-based processing).
- **Indirection table (stable handles AND dense data)** - Get both at once: a sparse `handle.index → data_index` map (with the generation), a dense data array iterated linearly, and a dense `data_index → handle.index` back-map. Deref is a one-hop lookup; remove is swap-remove on the dense array plus fixing the single moved element's map entry. Handles stay stable while iteration stays fully dense.

## Example

```c
typedef struct { uint32_t index, generation; } handle_t;

// Bad: raw pointer: dangles after the array reallocs or the slot is freed.
entity_t *cached = &items[i]; // grow() moves items[] -> cached now invalid

// Indirection table: stable handle + dense data, both. O(1) deref and remove.
typedef struct {
  entity_t *data;       // dense; iterate data[0..count)
  uint32_t *data_to_slot; // dense -> slot (back-map, for fixup on swap-remove)
  uint32_t *slot_to_data; // sparse handle.index -> dense index
  uint32_t *generation; uint32_t *free_list, free_count, count;
} store_t;

static entity_t *store_deref(store_t *s, handle_t h) {
  if (s->generation[h.index] != h.generation) return NULL;
  return &s->data[s->slot_to_data[h.index]]; // one hop, then dense
}
static void store_remove(store_t *s, handle_t h) {
  uint32_t d = s->slot_to_data[h.index], last = --s->count;
  s->data[d] = s->data[last];                 // swap-remove keeps data dense
  uint32_t moved_slot = s->data_to_slot[last];
  s->slot_to_data[moved_slot] = d;            // fix the ONE moved element's map
  s->data_to_slot[d] = moved_slot;
  s->generation[h.index]++; s->free_list[s->free_count++] = h.index;
}
```

## Gotchas

- Swap-remove invalidates the moved element's index; either use handles (which the move updates) or notify holders.
- Cache the dereferenced pointer only within a scope where no allocation/removal can occur.

## Related

[references/access-patterns.md](./access-patterns.md), [references/existence-based-processing.md](./existence-based-processing.md), memory-management-guide
