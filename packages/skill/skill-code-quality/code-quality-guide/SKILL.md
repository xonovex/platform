---
name: code-quality-guide
description: "Use when auditing existing code for quality WITHOUT changing it — a read-only pass that finds smells, grades them by severity, and routes each to its owner. Triggers on robustness / hardening (any-types, missing validation, swallowed errors, logging), duplication / dead code / over-abstraction / complexity, magic numbers, or a code-smell audit ('find issues in', 'is this robust', 'remove dead code', 'audit this module'). A smell catalog maps each smell to its owner."
---

# Code-Quality Audit

A read-only pass over existing code: find smells, grade them by severity, report — and route each smell to the one skill that owns its definition and fix. It changes nothing.

## Essentials

- **Read the project's own standards first** — `AGENTS.md` / `POLICY.md` / linked guidelines and the linter config decide what counts as a violation, not generic best-practice
- **Find → grade → report, read-only** — group findings by severity and category; applying the fix or authoring a plan is a separate step
- **Grade by blast radius × likelihood × remediation effort** — an unvalidated request body outranks a long pure helper; never grade by raw line count
- **Robustness is this skill's own dimension** — type safety, validation, error handling, logging at boundaries, complexity, see [references/robustness.md](references/robustness.md)
- **Route every other smell to its owner** — the catalog maps each smell to a detector signal and the skill that owns it, see [references/smell-catalog.md](references/smell-catalog.md)

## Gotchas

- This skill is a read-only **detector and grader** — it names a smell and routes to the owner (**oop-guide** for OO-design smells, **connascence-guide** for coupling smells); it does not redefine or fix them.
- A pile of `any` is usually a missing schema at one boundary, not a per-site typing problem — fix the boundary, not every call site.
- "Dead code" detection misses code reached via dynamic dispatch, reflection, or external entry points — verify before flagging for deletion.
- A single-implementation interface is not automatically over-engineering — it may exist for testability or a planned variant.
- A comment that captures a non-obvious _why_ (a workaround, an invariant, a caveat) is not noise — flag a comment only when a rename would say it better, it restates the code, or it narrates a plan/provenance.
- The "application / class / method-level" smell grouping is an informal label, not a citable taxonomy — the smell catalog groups by design-problem family instead.
- Log at boundaries and on error paths, validate at trust boundaries — "everywhere" is noise and wasted work, not robustness.

## Example

```
=== Quality Audit — src/parsers/ ===
DUPLICATES (high)      makeNullable() — 6 near-identical impls → one shared helper       [here]
DEAD CODE (high)       legacyParser() never called — src/parsers/legacy.ts:45           [here]
GOD OBJECT (med)       ParserManager: 28 methods, low cohesion → split                   [→ oop-guide]
SWALLOWED ERROR (high) empty catch in load() makes a failure look like success          [robustness]
Impact: ~450 lines removable; 1 SRP split; 1 error path to surface
(read-only findings — no edits made)
```

## Progressive Disclosure

- Read [references/robustness.md](references/robustness.md) - Load when auditing robustness: type safety, validation, error handling, logging, complexity
- Read [references/smell-catalog.md](references/smell-catalog.md) - Load to map a smell to its detector signal and owning skill (oop-guide / connascence-guide / here)
