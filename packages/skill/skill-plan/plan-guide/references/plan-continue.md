# plan-continue: Continue Progress from a Plan

Resume work from an existing plan document. **Start working immediately — no questions, no permission-asking, no confirmations.** **One plan only:** when complete, STOP; the user re-runs for the next plan. This is the only plan operation that modifies the codebase.

## Workflow

**Work on ONE plan at a time. Do NOT read all child plans upfront — metadata-only scan first, then load only the target.**

1. **Resolve plan** (if no path provided):
   - `git config branch.$(git branch --show-current).plan`
   - else conversation context (recent `plan-research` / `plan-create` results)
   - else ask the user. Do NOT glob unprompted.
2. **Load parent metadata only** — first ~30 lines for `has_subplans`, `status`, `parallel_group`
3. **Find next actionable child** (if `has_subplans: true`) — list `plans/<plan-name>/` child files, read only the first 30 lines of each in order (01, 02, 03…), stop at the FIRST with `status: in_progress` or `pending`
4. **Determine target** — parent `status: complete` → warn and exit · found actionable child → target · all children complete → recommend `plan-update` to mark parent complete · no children → parent is the target
5. **Load the target plan fully**; analyze phase, completed/pending tasks, blockers
6. **Baseline** — verify files exist, run typecheck / lint / build / tests, report state
7. **Consult skills** (hard gate — implementation must not begin until loaded) — read the plan frontmatter `skills_to_consult` and invoke EACH skill, reading its relevant `references/*.md` not just the `SKILL.md` summary
8. **Create tasks** for THIS plan's highest-priority pending work, then **execute** all of them
9. **On completion** — set plan `status: complete`, run validation, report summary, STOP

## Output

```
Continuing plan: plans/authentication.md
Progress: 2/4 child plans complete
Current: 03-api-routes.md (in_progress)
Validation:
  [PASS] Typecheck · [PASS] Lint · [PASS] Build
  [WARN] Tests: 2 failing (expected — not yet implemented)
```

## Gotchas

- Reading all child plans upfront wastes context — metadata-only scan first, then load only the target
- Skipping the baseline at start hides regressions from previous plans — always validate before working
- Auto-continuing to the next plan after completion silently chains work — STOP after one and let the user decide
- `skills_to_consult` must actually be invoked and each skill's relevant reference docs read — the `SKILL.md` summary alone often omits the detail
