# determinism-and-replay: Making Intermittent Bugs Reproducible

Convert an intermittent bug into a deterministic one before trying to fix it: a reproducible bug can be stepped through _before_ it fails; a random one forces logging-after-the-fact.

## How to Apply

1. Hunt nondeterminism: uninitialized memory, unordered iteration, time/random seeds, thread scheduling. Add a "run single-threaded" flag, if the bug vanishes single-threaded it's a race (reclamation detail in lock-free-guide).
2. Where feasible, make the system's tick a pure function of its inputs (same inputs → same state), so a captured input sequence reproduces the bug every time.
3. Still resists? Use a record/replay or reverse-execution tool to step _backward_ from the failure; verify the recorded run actually reproduces the bug.
4. At scale, take the statistical view: collect crash reports (stack trace + memory dump), group by stack-trace + message, prioritize by users affected × frequency.

## Example

```c
bool g_single_threaded = false;   // flip on: if the bug disappears, it's a race
```

### Gotchas

- Adding logging or a single-threaded flag can perturb timing and hide a race (heisenbug): vanishing under instrumentation is _evidence_ of a timing/ordering bug, not a fix.
- Record/replay tools have real overhead and limits on large threaded programs; treat as a fallback and verify the capture reproduces the failure.

### Related

[references/reproduction-and-bisection.md](./reproduction-and-bisection.md), [references/instrumentation-and-checks.md](./instrumentation-and-checks.md), **lock-free-guide**, **data-oriented-design-guide**
