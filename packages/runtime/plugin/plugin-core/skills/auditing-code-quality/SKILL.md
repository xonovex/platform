---
name: auditing-code-quality
description: "Use when auditing existing code for quality and reporting graded smells without changing anything. Triggers on code smells, robustness or hardening, duplication, dead code, over-abstraction, complexity, magic numbers, inconsistent implementations, barrel exports, redundant comments, a TODO/FIXME inventory, or 'report only, no edits', even when the user doesn't say 'audit' or 'code quality'."
---

# Code-Quality Audit

A read-only pass over existing code: find smells, grade them by severity, report, and route each smell to the one skill that owns its definition and fix. It changes nothing.

## Essentials

- **Read the project's own standards first**: `AGENTS.md` / `POLICY.md` / linked guidelines and the linter config decide what counts as a violation, not generic best-practice
- **Robustness is this skill's own dimension**: type safety, validation, error handling, logging at boundaries, complexity, see [references/robustness.md](references/robustness.md)
- **Route every other smell to its owner**: the catalog maps each smell to a detector signal and the skill that owns it, see [references/smell-catalog.md](references/smell-catalog.md)
- **Keep cleanup inventories read-only**: preserve functional comments and rationale; detect prose debt and other cleanup candidates, see [references/comment-inventory.md](references/comment-inventory.md) and [references/cleanup-inventories.md](references/cleanup-inventories.md)

## Gotchas

- This skill is a read-only **detector and grader**. It names a smell and routes to the owner (**applying-oop** for OO-design smells, **analyzing-connascence** for coupling smells); it does not redefine or fix them.
- "Dead code" detection misses code reached via dynamic dispatch, reflection, or external entry points: verify before flagging for deletion.
- A single-implementation interface is not automatically over-engineering. It may exist for testability or a planned variant.
- Preserve comments that explain a non-obvious why, invariant, workaround, or caveat; flag restatements and plan or provenance narration.
- The "application / class / method-level" smell grouping is an informal label, not a citable taxonomy: the smell catalog groups by design-problem family instead.
- Log at boundaries and on error paths, validate at trust boundaries: "everywhere" is noise and wasted work, not robustness.

## Progressive Disclosure

- Read [references/robustness.md](references/robustness.md) - Load when auditing robustness: type safety, validation, error handling, logging, complexity
- Read [references/smell-catalog.md](references/smell-catalog.md) - Load when mapping a smell to its detector signal and owning skill (**applying-oop** / **analyzing-connascence** / here)
- Read [references/comment-inventory.md](references/comment-inventory.md) - Load when inventorying redundant, stale, narrative, commented-out, directive-like, or ambiguous comments and docstrings
- Read [references/cleanup-inventories.md](references/cleanup-inventories.md) - Load when inventorying barrels, duplicated shared candidates, or TODO/FIXME clusters
