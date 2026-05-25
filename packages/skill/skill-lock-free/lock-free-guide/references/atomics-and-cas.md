# atomics-and-cas: Atomics and Compare-and-Swap

**Guideline:** Use the narrowest atomic primitive that expresses the operation; for conditional updates use a `compare_exchange_weak` retry loop that reloads the expected value on every failure.

**Rationale:** Atomic read-modify-write (RMW) ops are the only way to update shared state without a lock while keeping it consistent. A plain load-then-store is two operations and races; an RMW is one indivisible step. CAS is the universal primitive — it can build any other RMW — but it must be written as a loop because another thread may win the race between your read and your write.

**Primitives (`<stdatomic.h>`):**

- `atomic_load` / `atomic_store` — single indivisible read or write.
- `atomic_exchange` — store new, return old, atomically (an unconditional swap; the basis of the MPSC tail update).
- `atomic_fetch_add` / `_sub` / `_or` / `_and` / `_xor` — atomic arithmetic/bitwise, returning the prior value. Prefer these over a CAS loop when they fit: they are wait-free on most ISAs (x86 `lock xadd`), whereas a CAS loop is only lock-free.
- `atomic_compare_exchange_weak(obj, &expected, desired)` / `_strong` — if `*obj == expected`, set it to `desired` and return true; else write the current value into `expected` and return false.

**weak vs strong, spurious failure:**

- `compare_exchange_weak` may fail *even when `*obj == expected`\* (a spurious failure), because it compiles to a single LL/SC pair on ARM/POWER where an unrelated event can break the reservation. It is cheaper and is what you want inside a loop.
- `compare_exchange_strong` retries internally to suppress spurious failures. Use it only when you are _not_ already looping (a one-shot conditional update).

**The CAS loop idiom:**

```c
// Atomically apply f() to a shared value (e.g. clamp, max, bitset).
int atomic_apply(atomic_int *p, int (*f)(int)) {
    // relaxed load: we re-validate inside the CAS, so no ordering needed yet.
    int expected = atomic_load_explicit(p, memory_order_relaxed);
    int desired;
    do {
        desired = f(expected);
        // weak: spurious failure just spins us once more.
        // acq_rel on success: this RMW publishes/consumes around the update.
        // On failure, compare_exchange writes the current value back into
        // `expected`, so the next f() runs on fresh state — no manual reload.
    } while (!atomic_compare_exchange_weak_explicit(
                 p, &expected, desired,
                 memory_order_acq_rel, memory_order_acquire));
    return desired;
}
```

The failure-ordering argument (the last one) must not be stronger than the success ordering and must not be a release order — `memory_order_acquire` or `relaxed` are the usual choices.

**LL/SC vs CAS:** ARM/POWER expose Load-Linked / Store-Conditional, where SC fails if the line was touched since LL. This avoids ABA _within one LL/SC window_ but cannot span a function call, so portable code still uses CAS and must handle ABA separately. x86 offers `cmpxchg` (CAS) and `lock xadd` directly.

**When RMW is needed:** any time the new value depends on the current shared value (increment, push onto a list head, claim a slot). If the write is unconditional and value-independent, a plain `atomic_store` suffices.

**Related:** [references/aba-problem.md](./aba-problem.md), [references/memory-ordering.md](./memory-ordering.md), [references/lock-free-stack.md](./lock-free-stack.md)
