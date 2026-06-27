# testing-and-verification: Testing and Verification

## Guideline

Concurrency bugs are non-deterministic and rare; a single happy-path test proves nothing. Every lock-free structure must ship with a stress test and pass under a race detector.

## Rationale

A data race or missing fence may surface once in a billion interleavings, only on weak-memory hardware, only under load — exactly the conditions you don't reproduce on a dev box. Ordinary tests run one interleaving and pass. You need tools that _amplify_ rare interleavings (stress + randomized scheduling) and tools that _reason_ about all of them (sanitizers and model checkers).

## How to Apply

1. **Stress test:** run many threads (≥ 2× cores) hammering the structure for millions of operations, then check an invariant — e.g. every pushed item is popped exactly once, the final sum matches, no duplicate or lost element. Add randomized sleeps/`sched_yield` to widen interleavings.
2. **ThreadSanitizer (TSan):** build with `-fsanitize=thread` and run the stress test. TSan instruments memory accesses and flags data races (a missing `_Atomic`, a relaxed where you needed acquire). It catches the _race_, not just a wrong result — invaluable, but only finds races it actually observes, so run it on the stress test.
3. **Relacy Race Detector:** a C++ framework that exhaustively (or randomly) explores interleavings of a _modeled_ version of your algorithm under a relaxed memory model, including reorderings TSan can't see on your hardware. The standard tool for validating a new lock-free algorithm's memory ordering before porting it to production C.
4. **Model checkers (CDSChecker, GenMC):** systematically explore all executions of small C11-atomics programs under the formal memory model, proving the orderings correct (or producing a counterexample trace). Use on the core of a novel algorithm.
5. **Assertions:** assert structural invariants in debug builds (head/tail relationships, sequence numbers, refcounts ≥ 0). They turn a silent corruption into a loud, localized failure.

## Example

```c
// Stress harness: N producers + N consumers, verify conservation of items.
#define ITEMS_PER_PRODUCER 1000000
atomic_int produced_sum = 0, consumed_sum = 0;

void *producer(void *arg) {
    queue_t *q = arg;
    for (int i = 1; i <= ITEMS_PER_PRODUCER; i++) {
        while (!q_enqueue(q, i)) cpu_relax();        // retry on full
        atomic_fetch_add_explicit(&produced_sum, i, memory_order_relaxed);
    }
    return NULL;
}

void *consumer(void *arg) {
    queue_t *q = arg; int v;
    while (still_running()) {
        if (q_dequeue(q, &v))
            atomic_fetch_add_explicit(&consumed_sum, v, memory_order_relaxed);
        else cpu_relax();
    }
    return NULL;
}

// After joining all threads and draining:
assert(produced_sum == consumed_sum);  // nothing lost, nothing duplicated
```

## Gotchas

- "It passed 1000 times" is not a proof — absence of a TSan/Relacy finding is far stronger evidence than a green stress run.
- TSan does not model weak hardware reordering (it runs your code on x86's strong model); a queue that passes TSan can still break on ARM. Combine TSan (finds races) with Relacy/CDSChecker (finds ordering bugs).
- Make the harness deterministic to _replay_ a failure: seed the RNG that drives sleeps, and log the seed on assertion failure.

## Related

[references/memory-model.md](./memory-model.md), [references/memory-ordering.md](./memory-ordering.md)
