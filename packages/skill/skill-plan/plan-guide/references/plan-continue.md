# plan-continue: Continue One Planning Result

Resume implementation from one parent or child Planning result. Start immediately, complete only the first actionable result, publish status/evidence, and stop.

## Core workflow

1. **Resolve input** — use an explicit provider context + opaque native reference; otherwise current reconstructed handle, unambiguous workspace association, or ask. Never glob for an arbitrary plan or infer identity from a file name.
2. **Reconstruct** — ask the selected provider to resolve the exact native revision and rebuild the ephemeral `PhaseResultHandle`; verify freshness, available capabilities, profile/policy requirements, and status.
3. **Select one** — if the result relates child plans, read only provider-native metadata in declared order until the first `in_progress` or `pending` child. Do not load every child body.
4. **Load target** — resolve that one target fully, including tasks, dependencies, blockers, success criteria, source references, and skills/capabilities to consult.
5. **Baseline** — discover the actual toolchain beyond one manifest and run applicable typecheck, lint, build, tests, and integration checks before changes. Record unavailable categories explicitly.
6. **Consult skills** — load every target `skills_to_consult` capability and its relevant progressive references before implementation; fail visibly for a mandatory unavailable skill.
7. **Execute** — create task tracking for the target's highest-priority pending work, implement all of it, and keep Development evidence bound to the exact workspace/subject revision.
8. **Verify** — re-read tasks and success criteria, run applicable validation, and resolve warnings at root cause.
9. **Publish** — write complete/blocked status, validation evidence, limitations, and resulting exact revision through the selected provider; stop without advancing to another child.

Apply [early-lifecycle-contracts.md](early-lifecycle-contracts.md). Local plan files and Git worktrees are supported only when selected providers/workspace adapters own them.

## Gotchas

- Unchanged conversation memory is not fresh-context recovery; provider resolution is.
- Passing tests do not prove every success criterion or Definition of Done item.
- Automatically continuing to the next child silently chains authority and scope.
