# instrumentation-and-checks: Assertions, Invariants, and Sanitizers as Tripwires

**Guideline:** Build tripwires into the code so a bug announces itself at the fault site instead of corrupting state silently — assert preconditions and invariants, tag and account allocations, validate at API boundaries, and run sanitizers — but only add a check that surfaces the fault _earlier or where it is cheaper to act on_, never one that merely restates what a crash already told you.

**Rationale:** Most expensive native bugs are expensive because the symptom appears far from the cause: a buffer overflow corrupts a neighbor that crashes minutes later, a use-after-free is a clean run until the freed memory is touched. Tripwires collapse that distance. An assertion converts an "unexpected initial condition" into an immediate, located failure that also documents the precondition and assigns blame to the caller. Allocation accounting (file/line tags, per-system totals, zero-on-shutdown) turns an invisible leak into a checkable number. A page-guarding allocator turns a silent overwrite into an instant access violation at the exact bad instruction. Sanitizers (ASan/UBSan/TSan) are the same idea automated by the toolchain. The discipline is to spend the check budget where it shortens the gap between fault and detection — a validation pass that only reprints "this pointer was already freed," which the debugger showed you for free, adds cost without value.

## Contents

- Assertions for preconditions and invariants
- Allocation instrumentation: file/line tags, per-system totals, zero-on-shutdown
- Page-guarding (end-of-page) allocation to catch overwrites at the fault site
- Reference/lifetime tracking as an error-detection mechanism
- Sanitizers (ASan / UBSan / TSan) as automated tripwires
- When a check is worth its cost

**How to Apply:**

1. Assert every precondition and invariant you actually control at the boundary that owns it: `assert(idx < count)`, `assert(ptr != NULL)`. Treat the assert as executable documentation of the contract.
2. Instrument allocations: pass `__FILE__`/`__LINE__` into the allocator, total bytes per subsystem, and assert each subsystem's counter is zero when it shuts down — proper teardown plus accounting is how leaks become visible.
3. For overwrite-prone allocations, use an end-of-page allocator that places each allocation against a page boundary and unmaps the page on free; an out-of-bounds write or write-after-free then faults immediately at the offending code rather than silently corrupting metadata or a neighbor.
4. When chasing stale references, track liveness as detection (not ownership): keep non-owning references with `retain()`/`release()` counts tagged with `__FILE__`/`__LINE__`, and assert the count is zero when the object is destroyed — this reports the leak _at destruction time_, before any use-after-free, and the tag tells you which reference leaked.
5. Run sanitizers in CI and locally: ASan for overflow/use-after-free, UBSan for undefined behavior, TSan for data races (race-specific guidance is owned by lock-free-guide). They are tripwires the compiler maintains for you.
6. Before adding a check, ask what it surfaces that the existing crash/debugger does not. Keep it if it fires earlier, at lower cost to diagnose, or in an automated test; drop it if it only re-states a fact you already get from the `.dmp` or debugger.

**Example:**

```c
// Precondition assert: unexpected initial condition becomes an immediate, located failure.
assert(num_flags < MAX_FLAGS);

// Allocation tagging + per-system accounting -> leaks become a checkable number.
void *my_malloc(size_t n, const char *file, int line);   // records file/line and adds to a per-system total
#define MY_ALLOC(n) my_malloc((n), __FILE__, __LINE__)
void system_shutdown(system_o *s) { assert(s->alloc_bytes == 0); }  // must return to zero

// Reference tracking as DETECTION, not ownership: refs don't keep the object alive,
// they just let destroy() prove no stale references remain.
void retain(truth_ref_t *r, const char *file, int line);  // tag with __FILE__/__LINE__
void release(truth_ref_t *r);
void destroy_truth(truth_o *t) { assert(t->refcount == 0); /* names the leaking site if not */ }
```

**Gotchas:**

- A check that only echoes what the crash already revealed (e.g. "you used a freed pointer") costs runtime and code for no diagnostic gain — instrument to fire _earlier_ or in a test, not to narrate the crash.
- Asserts must guard facts you control; asserting a hope turns into debugging the assertion when the real world disagrees.
- Refcounting-for-liveness here is a _detection_ tool, not an ownership/GC scheme — making references keep objects alive would just convert crashes into leaks and subtler state-divergence bugs (ownership design is memory-management-guide; concurrent refcount/reclamation is lock-free-guide).
- Page-guarding allocators are heavy (a page or more per allocation) — enable them for the suspect subsystem or a debug build, not globally in shipping.
- Validation that is too expensive to leave on gets disabled and then rots; size each tripwire so it can actually stay enabled where it matters.

**Related:** [references/bug-taxonomy.md](./bug-taxonomy.md), [references/determinism-and-replay.md](./determinism-and-replay.md), **c99-opinionated-guide**, **memory-management-guide**, **lock-free-guide**
