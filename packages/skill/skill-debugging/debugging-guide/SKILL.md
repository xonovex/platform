---
name: debugging-guide
description: "Use when chasing a bug in native or low-level software: a crash, access violation, use-after-free, leak, intermittent/heisenbug, or 'works on my machine' failure, and when deciding how to prevent a whole class of bugs by design. Triggers on segfaults, callstacks/.dmp files, freed-memory fill patterns (0xdddddddd), git bisect, minimal repro, assertions, sanitizers (ASan/UBSan/TSan), determinism, and record/replay, even when the user doesn't say 'debugging'. Skip lifetime/ownership design (use memory-management-guide), concurrency correctness and TSan-for-races (use lock-free-guide), and profiling for speed (use data-oriented-design-guide)."
---

# Debugging Guidelines

Systematic, hypothesis-driven debugging of native/low-level software, plus designing each class of bug out of existence. The cheapest bug is the one made impossible; the next cheapest is the one that trips a tripwire the instant it happens. Lifetime/ownership design lives in memory-management-guide, race correctness in lock-free-guide, and speed profiling in data-oriented-design-guide; this skill is about finding bugs and preventing classes of them.

## Essentials

- **Classify the bug** - Each class (logic, uninitialized, leak, overwrite, race, design, intermittent) has a design choice that eliminates the whole class, see [references/bug-taxonomy.md](references/bug-taxonomy.md)
- **Hypothesis, not flailing** - Form one falsifiable guess, change one thing, predict, observe; never edit randomly, see [references/scientific-debugging.md](references/scientific-debugging.md)
- **Get a deterministic repro first** - A reliable repro is what lets you step through the "before"; shrink it small, see [references/reproduction-and-bisection.md](references/reproduction-and-bisection.md)
- **Trip the wire early** - Assertions, invariants, validation layers, and sanitizers turn silent corruption into a loud failure at the fault site, see [references/instrumentation-and-checks.md](references/instrumentation-and-checks.md)

## Find the cause

- **Bisect to the change** - With a 100% repro, search history for the commit that introduced the bug, see [references/reproduction-and-bisection.md](references/reproduction-and-bisection.md)
- **Read the callstack and the bytes** - Access violations, freed-memory fill patterns, and watch-window values steer the next hypothesis, see [references/scientific-debugging.md](references/scientific-debugging.md)
- **Confirm before fixing** - Verify the hypothesis (breakpoint/log) before you commit to a fix direction, see [references/scientific-debugging.md](references/scientific-debugging.md)

## Make bugs catchable

- **Prefer prevention by design** - Asserts on preconditions, end-of-page allocators, type-safe APIs, and "clean up everything" expose bugs that would otherwise hide, see [references/bug-taxonomy.md](references/bug-taxonomy.md)
- **Tame intermittency with determinism** - Remove nondeterminism, log to a circular buffer, or record/replay so the rare bug reproduces, see [references/determinism-and-replay.md](references/determinism-and-replay.md)
- **Instrument allocations and invariants** - Tag allocations with file/line, count references, validate at API entry, see [references/instrumentation-and-checks.md](references/instrumentation-and-checks.md)

## Gotchas

- "Access violation" / "segfault" is about an unmapped address, not permissions or threading — it almost always means a garbage or dangling pointer was dereferenced.
- A bug that lands "in the same spot every time" is a reasonable but unproven 100% repro; if it later "disappears," revise the assumption rather than trusting it.
- A freed-memory fill pattern (e.g. `0xdddddddd`) at a sane mapped address points to use-after-free, not a random overwrite — random writes rarely produce a clean repeating byte pattern.
- Bisect lies when the repro is flaky or GOOD/BAD is mis-tagged, and incremental builds can mislead — force clean rebuilds and only bisect a deterministic repro.
- Adding validation that only re-reports what the debugger already showed you (e.g. "this pointer was freed") buys nothing; instrument to surface the fault _earlier or where it is cheaper to act on_, not to restate it.
- Garbage collection / refcount-for-liveness doesn't fix a stale-reference bug — it converts a loud crash into a quiet leak plus subtler "two systems, two states" logic bugs.
- Reaching for "compiler bug" before ruling out your own undefined behavior is almost always wrong; test another compiler/opt level and inspect the assembly first.

## Progressive Disclosure

- Read [references/bug-taxonomy.md](references/bug-taxonomy.md) - Load when classifying a bug or choosing a design that eliminates a whole class (uninitialized, leak, overwrite, race, logic, design, intermittent)
- Read [references/scientific-debugging.md](references/scientific-debugging.md) - Load when you have a repro and need a disciplined process: hypothesis, one change, predict, observe, narrow
- Read [references/reproduction-and-bisection.md](references/reproduction-and-bisection.md) - Load when building a minimal deterministic repro, shrinking inputs, or bisecting history to the offending commit
- Read [references/instrumentation-and-checks.md](references/instrumentation-and-checks.md) - Load when adding assertions, invariants, allocation tagging, validation layers, or running sanitizers as tripwires
- Read [references/determinism-and-replay.md](references/determinism-and-replay.md) - Load when a bug is intermittent and you need to make execution deterministic, stress it, log to a circular buffer, or record/replay it
