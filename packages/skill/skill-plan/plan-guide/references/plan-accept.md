# plan-accept: Approve a Plan for Execution

Record the user's approval on a plan document so it can be expanded into subplans. Terminal decision — do NOT edit the plan's content or implement anything.

## Precondition

A plan with `status: pending-approval` (or `rejected`, when re-accepting after fixes). If feedback is still open, run `plan-revise` first.

## Core Workflow

1. **Locate plan** — user message, git config, or most recent `plans/*.md`
2. **Sanity check before approving** — confirm `skills_to_consult` is populated, `dependencies` are reconciled, success criteria exist, and no unresolved annotation markers remain (`NOTE:` / `FIXME:` / `>>>`). If any are missing, report them and STOP — do not approve a half-finished plan
3. **Set `status: approved`** and refresh the `updated` date; change nothing else
4. **Present** a one-line confirmation and the next step (`plan-subplans-create`), then STOP

## Gotchas

- Approving only flips the status — it does not generate subplans or implement; `plan-subplans-create` is the next step and requires `status: approved`
- Never approve a plan with unresolved annotations or missing `skills_to_consult` — fix via `plan-revise` first
