# plan-continue: Continue Progress from a Plan

Resume work from an existing plan document. Auto-detect the associated plan from git config when available.

## Goal

**Start working immediately. No questions, no permission-asking, no confirmations.**

**One plan only.** When complete, STOP. The user re-runs the operation for the next plan.

- Auto-detect plan from git config or use a provided path
- Find first actionable child plan (if parent has subplans)
- Execute that single plan to completion
- Mark complete, report success, STOP

## Core Workflow

1. **Resolve plan** — check git config, then conversation context (do NOT search for plans)
2. **Load parent plan metadata** — read only the frontmatter / first 30 lines
3. **Find next actionable child plan** — first child with `pending` or `in_progress` status
4. **Load and execute** — read the target plan fully and implement it
5. **Complete and STOP** — mark plan complete, update status, report success

## Implementation Steps

**Work on ONE plan at a time. Do NOT read all child plans upfront.**

1. **Resolve plan** (if no path provided):
   - `git config branch.$(git branch --show-current).plan`
   - If no git config, use conversation context (recent `plan-research` / `plan-create` results)
   - Only if no context: ask user. Do NOT glob unprompted.
2. **Load parent plan metadata-only** — first ~30 lines for `has_subplans`, `status`, `parallel_group`
3. **Find next actionable child plan** (if `has_subplans: true`):
   - List child plan files in `plans/<plan-name>/`
   - For each in order (01, 02, 03…), read only first 30 lines for status
   - Stop at FIRST plan with `status: "in_progress"` or `pending`
4. **Determine work target**:
   - Parent `status: "complete"` → warn and exit
   - Found actionable child → that's the target
   - All children complete → recommend `plan-update` to mark parent complete
   - No children → parent is the target
5. **Load work target plan** (full)
6. **Analyze current status** — phase, completed tasks, pending tasks, blockers
7. **Validate state** — verify files exist, run typecheck / lint / build / tests, report baseline
8. **Consult skills** — before writing any code, read the plan frontmatter `skills_to_consult` and invoke EACH listed skill. For each, also read the relevant progressive-disclosure documents it points to (its `references/*.md`), not just the `SKILL.md` summary. Hard gate: implementation must not begin until they're loaded.
9. **Generate next steps** — highest priority pending tasks for THIS plan only
10. **Create tasks** — set up an actionable task list with status tracking
11. **Execute** — implement all tasks
12. **On completion** — set plan `status: complete`, run validation, report summary, STOP

## Output

**Initial briefing:**

```
Continuing plan: plans/authentication.md

Plan: Authentication System
Progress: 2/4 child plans complete
Current: 03-api-routes.md (in_progress)

Validation:
  [PASS] Typecheck: PASS
  [PASS] Lint: PASS
  [PASS] Build: PASS
  [WARN] Tests: 2 failing (expected — not yet implemented)
```

**Completion report:**

```
Completed plan: plans/authentication/03-api-routes.md

Status: in_progress → complete
Parent: plans/authentication.md (3/4 complete)
```

## Error Handling

- No git config + no conversation context → ask user
- Specified path doesn't exist → error
- Plan `status: "complete"` → warn (recommend `plan-validate` / `plan-update`)
- All child plans complete → info (recommend `plan-update` for parent)
- Malformed frontmatter → error

## Gotchas

- Reading all child plans upfront wastes context — metadata-only scan first, then load only the target
- Skipping the validation step at start hides regressions introduced by previous plans — always baseline before working
- Auto-continuing to the next plan after completion silently chains work — STOP after one plan and let the user decide
- `skills_to_consult` listed in the plan must actually be invoked, and each skill's relevant reference docs read — they encode the project conventions the implementation must follow, and the `SKILL.md` summary alone often omits the detail
