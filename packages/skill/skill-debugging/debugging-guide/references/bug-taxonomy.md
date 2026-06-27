# bug-taxonomy: Classes of Bugs and the Design That Eliminates Each

## Guideline

Treat every bug as a member of a class, and for each class prefer a design change that makes the whole class impossible (or trivially detectable) over fixing the one instance — classify first, then ask "what would have prevented every bug like this?"

## Rationale

Bugs are not all the same animal; the cheapest fix differs per class, and the highest leverage move is usually structural, not local. A logic error is 100% reproducible and yields to readable code; an uninitialized-condition bug yields to an assertion; a leak yields to allocation accounting; an overwrite yields to a page-guarding allocator; a race yields to a single-threaded fallback and a thread sanitizer; a design flaw yields only to a redesign. Spending a moment after each fix to identify the class and the structural prevention turns one fix into immunity from a family of future bugs — and "clean up / shut down properly" is itself a prevention technique, because tearing things down exposes lifetime bugs that lazy never-free code would hide.

## Contents

- The typo and the logic error — readability and fewer code paths
- The unexpected initial condition — assert preconditions
- The leak — account allocations per system, expect zero on shutdown
- The overwrite (buffer overrun / use-after-free) — end-of-page allocator
- The race — single-thread fallback + thread sanitizer (owned by lock-free-guide)
- The design flaw and the failed specification — type-safe, single-purpose APIs
- The intermittent / third-party / compiler bug — pointers to other techniques

### How to Apply

1. **Typo / logic error:** lower the surface area for the brain to "auto-correct" wrong code — descriptive names over `i`/`j`, `const` on values that never change, compiler warnings on (`-Wshadow`), a formatter, and fewer code paths (linear flow, shared idioms/macros for repetitive patterns like array erase). These are 100% reproducible, so they fall to a debugger and a careful read.
2. **Unexpected initial condition:** make expectations explicit with an assertion at the boundary — `assert(count < MAX)` documents the precondition and pins responsibility on the caller (see instrumentation-and-checks.md).
3. **Leak:** instrument allocations with `file`/`line`, total memory per subsystem through a per-system allocator, and assert the counter is back to zero when that system shuts down.
4. **Overwrite:** route suspect allocations through an end-of-page allocator so an out-of-bounds write or a write-after-free faults _immediately_ at the offending instruction instead of silently corrupting a neighbor.
5. **Race:** add a "run single-threaded" flag to confirm whether threading is involved, keep concurrency to simple well-known patterns, and run a thread sanitizer — race detection and reclamation are owned by lock-free-guide.
6. **Design flaw / failed spec:** step back and surface the unstated assumption; redesign so misuse is unrepresentable — single-purpose APIs whose behavior does not flip on a flag, types that prevent wrong calls, and explicit handles instead of implicit global state.

### Example

```c
// Unexpected initial condition -> assert the precondition (caller's responsibility).
void add_flags(flags_t *f, const flag_t *src, uint32_t n) {
    assert(f->count + n < MAX_FLAGS);   // documents + enforces the expectation
    memcpy(f->items + f->count, src, n * sizeof *src);
    f->count += n;
}

// Design flaw: ambiguous API. ensure_html_encoded() can't tell encoded from raw input.
void ensure_html_encoded(char *s);            // double-encodes or misses; unfixable as written
// Fix the CLASS: make the type carry the state so misuse is unrepresentable.
typedef struct { char *raw; } raw_text_t;
typedef struct { char *html; } html_text_t;
html_text_t html_encode(raw_text_t in);       // can only be called on raw text
```

### Gotchas

- "Never free / never shut down" looks simpler but _hides_ lifetime bugs — proper teardown is what exposes the stale-pointer and double-ownership bugs early, when they are cheap.
- An assertion encodes a belief; a wrong assertion sends you debugging the check instead of the code, so assert facts you actually control, not hopes.
- A design flaw cannot be coded around — recognizing the class saves you from "fixing" the same symptom repeatedly in different call sites.
- Reaching for "compiler bug" is the last resort: rule out your own undefined behavior, then test another compiler/optimization level and read the generated assembly before blaming the toolchain.

### Related

[references/instrumentation-and-checks.md](./instrumentation-and-checks.md), [references/determinism-and-replay.md](./determinism-and-replay.md), **memory-management-guide**, **lock-free-guide**, **c99-opinionated-guide**
