# Sources

## Game-engine development blog (archive)

- **URL:** https://archive-host.github.io/blog_archive/
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Essentials, Find the cause, Make bugs catchable, Gotchas
  - The framing of debugging as classifying bugs and preventing whole classes by design, and of debugging as a hypothesis-driven process anchored on a reliable reproduction.
- **Aspects extracted:**
  - "A Taxonomy of Bugs" — the bug classes (typo, logical error, unexpected initial condition, memory leak, memory overwrite, race condition, design flaw, third-party bug, failed specification, hard-to-reproduce bug, the statistic, compiler bug) and the per-class prevention: `-Wshadow`/formatter/`const`/descriptive names, fewer code paths, precondition `assert`, file/line allocation tagging + per-system totals + zero-on-shutdown, end-of-page allocator, single-threaded fallback + thread sanitizer, type-safe single-purpose APIs, multi-compiler/opt-level + reading assembly for suspected compiler bugs → `references/bug-taxonomy.md` and `references/instrumentation-and-checks.md`
  - "A Debugging Story" — the default strategy of breaking in _before_ the bug on a reliable repro; ranking hypotheses (access violation = unmapped address, not permissions/threading; crash on first line rules out in-function logic; `0xdddddddd` at a mapped address ⇒ use-after-free not overwrite); confirming the hypothesis with a breakpoint/log before fixing; verifying the fix → `references/scientific-debugging.md`
  - "A Debugging Story" — simplifying the repro from a slow packaging task to "start app, New Project twice"; treating "100% reproducible" as a revisable assumption; using `git bisect` on a deterministic repro and its failure modes (flaky repro, mislabeled GOOD/BAD, stale incremental builds, slow); the offending commit points to a subsystem, not a line → `references/reproduction-and-bisection.md`
  - "A Debugging Story" + "A Taxonomy of Bugs" — increasing reproduction rate by stressing the failing system; circular in-memory logging dumped on detection; `rr`-style record/replay and step-back-in-time; the single-threaded fallback flag to isolate races; automatic crash reporting grouped by stack-trace + message with memory dumps for the statistical view → `references/determinism-and-replay.md`
  - "A Debugging Story" — prevention reflections: "clean up / shut down properly exposes hidden bugs" vs never-free; validation that only restates the crash adds no value; GC/keep-alive refcount converts crashes into leaks + state-divergence bugs; reference _counting for detection_ (non-owning refs, `retain`/`release`, `__FILE__`/`__LINE__` tags, assert zero at destroy) → `references/instrumentation-and-checks.md` and `references/bug-taxonomy.md`

## Reverse-execution / record-replay tooling

- **URLs:**
  - https://rr-project.org/
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Make bugs catchable
- **Aspects extracted:** Record/replay and reverse-execution to step backward from a failure to its cause, and its overhead/limits on large programs → `references/determinism-and-replay.md`

## Sanitizers (compiler-maintained tripwires)

- **URLs:**
  - https://clang.llvm.org/docs/AddressSanitizer.html
  - https://clang.llvm.org/docs/UndefinedBehaviorSanitizer.html
  - https://clang.llvm.org/docs/ThreadSanitizer.html
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Make bugs catchable, Gotchas
- **Aspects extracted:** ASan (overflow/use-after-free), UBSan (undefined behavior), TSan (data races) as automated tripwires; race-specific guidance deferred to lock-free-guide → `references/instrumentation-and-checks.md`

## Refresh Workflow

1. Re-read the upstream source(s) above
2. Diff against the prior pull (or scan for newly added sections)
3. For each changed area, update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
