# determinism-and-replay: Making Intermittent Bugs Reproducible

Convert an intermittent bug into a deterministic one before trying to fix it — a reproducible bug can be stepped through _before_ it fails; a random one forces logging-after-the-fact.

## How to Apply

1. Hunt nondeterminism: uninitialized memory, unordered iteration, time/random seeds, thread scheduling. Add a "run single-threaded" flag — if the bug vanishes single-threaded it's a race (reclamation detail in lock-free-guide).
2. Where feasible, make the system's tick a pure function of its inputs (same inputs → same state), so a captured input sequence reproduces the bug every time.
3. Can't reproduce on demand? Raise the rate by stressing the suspect path — e.g. open/close 200 windows per frame to turn a once-an-hour crash into a few-seconds crash.
4. Capture the "before" without a live debugger: log key events to a fixed-size circular buffer and dump it the instant the bug is detected.
5. Still resists? Use a record/replay or reverse-execution tool to step _backward_ from the failure; verify the recorded run actually reproduces the bug.
6. At scale, take the statistical view: collect crash reports (stack trace + memory dump), group by stack-trace + message, prioritize by users affected × frequency.

## Example

```c
bool g_single_threaded = false;   // flip on: if the bug disappears, it's a race

for (int i = 0; i < 200; ++i) { open_window(); close_window(); }  // stress, per frame

// Circular log: keep only recent history, dump it when the bug trips.
typedef struct { char lines[1024][128]; uint32_t head; } ring_log_t;
void log_event(ring_log_t *l, const char *msg);       // overwrites oldest; O(1), bounded
void on_bug_detected(ring_log_t *l) { dump_ring(l); } // emit the "before" at failure
```

### Gotchas

- Adding logging or a single-threaded flag can perturb timing and hide a race (heisenbug) — vanishing under instrumentation is _evidence_ of a timing/ordering bug, not a fix.
- A circular buffer that's too small drops the events you needed; size it to cover earliest plausible cause → detection.
- Stressing changes allocation patterns and ordering too — confirm the stressed crash is the _same_ bug, not a new one you provoked.
- Record/replay tools have real overhead and limits on large threaded programs; treat as a fallback and verify the capture reproduces the failure.

### Related

[references/reproduction-and-bisection.md](./reproduction-and-bisection.md), [references/instrumentation-and-checks.md](./instrumentation-and-checks.md), **lock-free-guide**, **data-oriented-design-guide**
