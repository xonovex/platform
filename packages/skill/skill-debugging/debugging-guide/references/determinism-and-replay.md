# determinism-and-replay: Making Intermittent Bugs Reproducible

**Guideline:** When a bug comes and goes, make it reproducible before trying to fix it — remove sources of nondeterminism, stress the suspect system to raise the hit rate, capture state into a circular log that dumps on detection, or record/replay execution to step through the "before" after the fact.

**Rationale:** A reproducible bug can be stepped through in a debugger _before_ it fails; a random one cannot, because you don't know in advance which run will break — so you're forced into logging-after-the-fact or time-travel tooling. The whole battle with intermittent bugs is converting them into deterministic ones. Determinism is also a property you can design _in_: an engine whose update is a pure function of its inputs reproduces the same bug on the same inputs every time, which makes intermittent failures into ordinary ones. When you can't get a clean repro, you fall back to raising the reproduction rate (stress) and to capturing enough state at the moment of failure (circular logs, crash dumps) to reconstruct the cause without a live debugger.

## Contents

- Why determinism matters: reproducibility enables "step through the before"
- Designing for determinism: execution as a function of inputs
- Raising the reproduction rate by stressing the failing system
- Circular logging that dumps on detection
- Record/replay and reverse-execution tooling
- Crash dumps and the statistical view at scale

**How to Apply:**

1. Hunt down nondeterminism: uninitialized memory, unordered iteration, time/random seeds, and thread scheduling all make runs differ. Confirm whether threading is the variable by adding a "run single-threaded" flag — if the bug vanishes single-threaded, it's a race (reclamation/race detail in lock-free-guide).
2. Where feasible, make the system's tick a deterministic function of its inputs (same inputs → same state), so a captured input sequence reproduces the bug every time.
3. If you can't reproduce on demand, raise the rate by stressing the suspect path — e.g. open and close 200 windows per frame to turn a once-an-hour crash into a few-seconds crash.
4. Capture the "before" without a live debugger: log key events to a fixed-size circular buffer in memory and dump it the instant the bug is detected, so you keep the recent history that led to the failure without the cost of logging everything to disk.
5. For bugs that still resist, use a record/replay or reverse-execution tool to capture a run and step _backward_ from the failure to the cause; understand it can be heavy for large programs and verify the recorded run actually reproduces the bug.
6. At scale, accept the statistical view: collect automatic crash reports with stack traces and memory dumps, group by stack-trace + message, and prioritize by how many users each bug affects and how often — the dump lets you inspect state after the fact.

**Example:**

```c
// 1) Isolate threading as the variable: a single-threaded fallback for triage.
bool g_single_threaded = false;   // flip on to test: if the bug disappears, it's a race

// 2) Stress to raise a low reproduction rate into a fast, reliable one.
for (int i = 0; i < 200; ++i) { open_window(); close_window(); }  // per frame

// 3) Circular log: keep only the recent history, dump it when the bug trips.
typedef struct { char lines[1024][128]; uint32_t head; } ring_log_t;
void log_event(ring_log_t *l, const char *msg);     // overwrites oldest; O(1), bounded memory
void on_bug_detected(ring_log_t *l) { dump_ring(l); } // emit the "before" at the moment of failure
```

**Gotchas:**

- Adding logging or a single-threaded flag can perturb timing and hide a race (a heisenbug) — when it vanishes under instrumentation, that itself is evidence of a timing/ordering bug, not a fix.
- A circular buffer that's too small drops the events you needed; size it to cover the window between the earliest plausible cause and the detection point.
- Stressing changes more than frequency — it can also change allocation patterns and ordering, so confirm the stressed crash is the _same_ bug, not a new one you provoked.
- Record/replay and reverse-execution tools have real overhead and limits on large, threaded programs; treat them as a fallback for bugs that defeat stress + logging, and verify the capture reproduces the failure before trusting it.
- "Make it deterministic" is not free: a system that depends on wall-clock time, addresses, or thread interleaving needs design changes to become replayable — budget for that, don't assume it.

**Related:** [references/reproduction-and-bisection.md](./reproduction-and-bisection.md), [references/instrumentation-and-checks.md](./instrumentation-and-checks.md), **lock-free-guide**, **data-oriented-design-guide**
