# bug-taxonomy: Classes of Bugs and the Design That Eliminates Each

Classify every bug into a class, then prefer the structural change that makes the whole class impossible (or trivially detectable) over fixing the one instance. Ask "what would have prevented every bug like this?"

## Class → prevention

1. **Typo / logic error** (100% reproducible): descriptive names over `i`/`j`, `const` on invariant values, `-Wshadow` and warnings on, a formatter, fewer code paths. Falls to a debugger + careful read.
2. **Unexpected initial condition:** assert the precondition at the boundary that owns it — `assert(count < MAX)` documents it and pins blame on the caller.
3. **Leak:** tag allocations with `__FILE__`/`__LINE__`, total bytes per subsystem, assert the counter is zero on that system's shutdown.
4. **Overwrite (buffer overrun / use-after-free):** route suspect allocations through an end-of-page allocator so an OOB or after-free write faults _immediately_ at the offending instruction.
5. **Race:** add a "run single-threaded" flag to confirm threading is involved, keep concurrency to known patterns, run TSan — reclamation/race detail owned by lock-free-guide.
6. **Design flaw / failed spec:** redesign so misuse is unrepresentable — single-purpose APIs that don't flip behavior on a flag, types that prevent wrong calls, explicit handles over global state.

Proper teardown ("clean up everything") is itself a prevention technique: tearing down exposes lifetime bugs that never-free code hides.

## Example

```c
// Unexpected initial condition -> assert the precondition (caller's responsibility).
void add_flags(flags_t *f, const flag_t *src, uint32_t n) {
    assert(f->count + n < MAX_FLAGS);
    memcpy(f->items + f->count, src, n * sizeof *src);
    f->count += n;
}

// Design flaw: ambiguous API can't tell encoded from raw input.
void ensure_html_encoded(char *s);            // double-encodes or misses; unfixable as written
// Fix the CLASS: make the type carry the state so misuse is unrepresentable.
typedef struct { char *raw; } raw_text_t;
typedef struct { char *html; } html_text_t;
html_text_t html_encode(raw_text_t in);       // can only be called on raw text
```

### Gotchas

- "Never free / never shut down" _hides_ lifetime bugs — proper teardown exposes stale-pointer and double-ownership bugs early, when cheap.
- An assertion encodes a belief; assert facts you control, not hopes, or you end up debugging the check.
- A design flaw cannot be coded around — recognizing the class saves "fixing" the same symptom in different call sites.
- "Compiler bug" is the last resort: rule out your own UB, test another compiler/optimization level, and read the generated assembly first.

### Related

[references/instrumentation-and-checks.md](./instrumentation-and-checks.md), [references/determinism-and-replay.md](./determinism-and-replay.md), **memory-management-guide**, **lock-free-guide**, **c99-opinionated-guide**
