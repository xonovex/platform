# virtual-memory: Virtual-Memory Tricks Beyond Reserve/Commit

## Guideline

Treat address-space reservation as nearly free (especially on 64-bit) and exploit it: back cap-free arrays by reserving a huge range and committing on touch, grow buffers in page multiples, alias a second mapping for a gapless ring buffer, and use an end-of-page allocator to catch overruns. For the reserve/commit basics see [arenas-and-pools.md](./arenas-and-pools.md); this is the toolbox built on top.

## Rationale

Reserving address space only consumes page-table entries, not physical RAM — and on 64-bit the address space (~tens of TB per process) is effectively unlimited, so virtual fragmentation is irrelevant. That decouples "how big could this get" from "how much memory does it use," which removes static caps, the realloc-and-copy on growth, and the split-on-wrap logic in ring buffers. The same mechanism gives a precise, hardware-enforced bounds checker for free.

## How to Apply

1. **Cap-free arrays** - Reserve `MAX * sizeof(elem)` up front (e.g. 8 GB) but commit pages only as the array grows; the array never moves, so pointers into it stay valid (no realloc copy). Physical cost is only the touched pages.
2. **Page-aligned growth** - Grow by whole pages (4K, 8K, …) instead of geometric element doubling; this cuts average internal fragmentation from ~half-a-doubling down to a fraction of a page, while staying amortized O(1).
3. **Gapless ring buffer** - Map a second virtual region immediately after the buffer onto the _same_ physical pages; wraparound reads/writes then need no split logic because the alias makes the end contiguous with the start.
4. **End-of-page allocator (debug)** - Place the allocation's end flush against a page boundary with the next page unmapped, so any overrun faults immediately. Use only while hunting corruption; revert for release.
5. Reserve everything, commit the working set — `MEM_COMMIT` (or its equivalent) can still fail if the page file/backing store is too small even when RAM is free.

## Example

```c
// Cap-free, never-moving array: reserve 8 GB of address space, commit on demand.
elem_t *items = vm_reserve(MAX_ITEMS * sizeof(elem_t));      // address space only
void push(elem_t e) {
    if ((char *)&items[count] >= committed_end)
        committed_end = vm_commit(committed_end, PAGE_SIZE);  // grow by a page, no copy
    items[count++] = e;                                       // &items[i] stays valid forever
}

// End-of-page bounds checker (debug): object end == page end, next page unmapped.
void *eop_malloc(uint64_t size) {
    uint64_t pages = (size + PAGE_SIZE - 1) / PAGE_SIZE;
    char *base = vm_reserve_commit(pages * PAGE_SIZE);
    return base + (pages * PAGE_SIZE - size);                 // overrun -> immediate fault
}
```

## Counter-Example

On 32-bit (or tightly-constrained embedded targets) address space is scarce and reservations are not free — there, large reservations and page-aligned growth trade away precious address range, so the classic capped/geometric approaches win. The tricks assume a roomy 64-bit address space.

## Gotchas

- Linux overcommits, so a reserve and a commit often look the same; Windows separates `MEM_RESERVE` from `MEM_COMMIT` — write to the OS contract you target.
- The ring-buffer double-mapping can race: there's no guarantee the region right after your buffer stays free between reserving and mapping it — reserve, free, then map the alias and retry if it was taken.
- After freeing pages in the end-of-page allocator, reuse can mask a use-after-free; keep freed pages reserved-but-uncommitted so a later access still faults.
- Page-aligned growth trades physical memory for reduced internal fragmentation — fine on modern hardware, costly on memory-tight platforms.

## Related

[arenas-and-pools.md](./arenas-and-pools.md), [ownership-and-lifetimes.md](./ownership-and-lifetimes.md); **data-oriented-design-guide** for handle/index storage over a stable-address block
