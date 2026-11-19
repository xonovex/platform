---
name: approving-plan
description: "Use when recording the user's approval or rejection of a reviewed plan. Triggers on accepting a plan after a final sanity check, marking a plan approved, rejecting a plan with a recorded reason, a plan that waits for sign-off, and the question whether a plan is ready to approve, even when the user doesn't say 'approve' and only says 'this plan is fine by me' or 'no, not this plan'."
---

# Approve or Reject a Plan

Record the user's decision on a reviewed plan: approval, or rejection with a reason. The decision is a record of the review. It leaves the rest of the plan as it is, and it implements nothing.

## Inputs

- **Plan** (optional): the plan document; taken from the user message, the git configuration, or the most recent plan in `plans/` when absent
- **Decision** (required): approve or reject, as the request states it
- **Reason** (optional): why the plan is rejected; asked for when absent

## Approve

Precondition: a plan with `status: pending-approval` (or `rejected`, when re-accepting after fixes). If feedback is still open, apply it with **revising-plan** first.

1. **Locate plan**, user message, git config, or most recent `plans/*.md`
2. **Sanity check before approving**, confirm `skills_to_consult` is populated, `dependencies` are reconciled, success criteria exist, and no unresolved annotation markers remain (`NOTE:` / `FIXME:` / `>>>`). If any are missing, report them and STOP, do not approve a half-finished plan
3. **Set `status: approved`** and refresh the `updated` date; change nothing else
4. **Present** a one-line confirmation and the next step (**defining-subplans** or **implementing-plan**), then STOP

## Reject

The plan is not deleted: it stays for revision or reference.

1. **Locate plan**, user message, git config, or most recent `plans/*.md`
2. **Capture the reason**, from the prompt; if none was given, ask the user why in one sentence. A rejection without a reason is not actionable
3. **Record it**, set `status: rejected`, refresh the `updated` date, and append the reason under a `## Rejection` section (date + reason) so the history stays visible
4. **Present** the reason and the suggested next step, **revising-plan** to address it, **researching-changes** if context was missing, or leave it rejected, then STOP

## Boundaries

An independent review before the decision belongs to **critiquing-plan**. Changes to the plan content belong to **revising-plan**.

## Gotchas

- Approval only sets the status: it generates no subplans and implements nothing
- Never approve a plan with unresolved annotations or missing `skills_to_consult`, fix them through **revising-plan** first
- Never delete the plan on rejection, the reason and history are the point
- A rejected plan stays open for revision and a new decision
- Status is optional descriptive metadata: do not use approval or status fields to authorize or gate another planning skill

## Progressive Disclosure

- Read [references/example.md](references/example.md) - Load when wording the confirmation of an approval or a rejection
