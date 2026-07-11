# instrumentation-and-checks: Assertions, Invariants, and Sanitizers as Tripwires

Build tripwires so a bug announces itself at the fault site instead of corrupting state silently — but only add a check that surfaces the fault _earlier or where it is cheaper to act on_, never one that merely restates what a crash already told you.

## How to Apply

1. Assert every precondition/invariant you control at the owning boundary: `assert(idx < count)`, `assert(ptr != NULL)`. The assert is executable documentation of the contract.
2. Instrument allocations: pass `__FILE__`/`__LINE__` into the allocator, total bytes per subsystem, assert each counter is zero at that subsystem's shutdown.
3. For overwrite-prone allocations, use an end-of-page allocator that places each allocation against a page boundary and unmaps the page on free — an OOB or after-free write then faults immediately at the offending instruction.
4. Chasing stale references? Track liveness as _detection_, not ownership: non-owning references with `retain()`/`release()` counts tagged `__FILE__`/`__LINE__`, assert the count is zero at destroy — reports the leak _at destruction time_, before any use-after-free, and names the leaking reference.
5. Run sanitizers in CI and locally: ASan (overflow/use-after-free), UBSan (undefined behavior), TSan (data races; race guidance owned by lock-free-guide).
6. Before adding a check, ask what it surfaces beyond the existing crash/debugger. Keep it if it fires earlier, cheaper to diagnose, or in an automated test; drop it if it only re-states a fact the `.dmp`/debugger already gives you.

## Example

```c
assert(num_flags < MAX_FLAGS);   // unexpected initial condition -> immediate, located failure

// Allocation tagging + per-system accounting -> leaks become a checkable number.
void *my_malloc(size_t n, const char *file, int line);
#define MY_ALLOC(n) my_malloc((n), __FILE__, __LINE__)
void system_shutdown(system_o *s) { assert(s->alloc_bytes == 0); }

// Reference tracking as DETECTION, not ownership: refs don't keep the object alive,
// they let destroy() prove no stale references remain.
void retain(truth_ref_t *r, const char *file, int line);
void release(truth_ref_t *r);
void destroy_truth(truth_o *t) { assert(t->refcount == 0); }  // names the leaking site if not
```

### Gotchas

- A check that only echoes the crash ("you used a freed pointer") costs runtime for no gain — instrument to fire _earlier_ or in a test.
- Asserts must guard facts you control; asserting a hope turns into debugging the assertion.
- Refcounting here is _detection_, not GC — keeping references alive would convert crashes into leaks and state-divergence bugs (ownership → memory-management-guide; concurrent refcount → lock-free-guide).
- Page-guarding allocators are heavy (≥1 page per allocation) — enable for the suspect subsystem or a debug build, not globally in shipping.
- Validation too expensive to leave on gets disabled and rots; size each tripwire so it can stay enabled where it matters.

### Related

[references/bug-taxonomy.md](./bug-taxonomy.md), [references/determinism-and-replay.md](./determinism-and-replay.md), **c99-opinionated-guide**, **memory-management-guide**, **lock-free-guide**
