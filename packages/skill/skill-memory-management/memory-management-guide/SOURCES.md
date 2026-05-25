# Sources

## Region-based / arena memory management

- **URL:** https://en.wikipedia.org/wiki/Region-based_memory_management and Tofte & Talpin, "Region-Based Memory Management" (Information and Computation, 1997)
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Allocation strategy, Ownership
  - Lifetime-by-reset and arena semantics
- **Aspects extracted:**
  - Arena/bump/linear allocators, lifetime reset, scratch arenas → `references/arenas-and-pools.md`

## Custom allocators (pools, freelists, bump)

- **URL:** Game Engine Architecture (J. Gregory), Memory Management chapter; Andrei Alexandrescu, "std::allocator Is to Allocation what std::vector Is to Vexation" (CppCon 2015)
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Allocation strategy
- **Aspects extracted:**
  - Object pools, free lists, fragmentation, per-allocation overhead → `references/arenas-and-pools.md`
  - Caller-provided storage / non-allocating APIs → `references/caller-owns-memory.md`

## Virtual memory reserve/commit

- **URL:** OS virtual-memory APIs — `mmap`/`mprotect` (POSIX), `VirtualAlloc` (Windows); reserve-then-commit growable-array technique
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Allocation strategy
- **Aspects extracted:**
  - Reserve large / commit on demand, stable-address growable arrays → `references/arenas-and-pools.md`

## Ownership & lifetimes

- **URL:** Ownership models as in RAII (C++) and the Rust ownership/borrow model — https://doc.rust-lang.org/book/ch04-00-understanding-ownership.html
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Ownership, Essentials
- **Aspects extracted:**
  - Single-owner/borrow, lifetime-by-scope, leak/double-free/use-after-free avoidance, refcount as exception → `references/ownership-and-lifetimes.md`

## Refresh Workflow

1. Re-read the upstream source(s) above
2. Diff against the prior pull (or scan for newly added sections)
3. For each changed area, update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
