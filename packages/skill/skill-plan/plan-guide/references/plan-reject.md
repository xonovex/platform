# plan-reject: Reject a Plan

Record that a plan is not approved, with a reason, so it is not expanded or executed. Does NOT delete the plan — it stays for revision or reference.

## Core Workflow

1. **Locate plan** — user message, git config, or most recent `plans/*.md`
2. **Capture the reason** — from the prompt; if none was given, ask the user why in one sentence. A rejection without a reason is not actionable
3. **Record it** — set `status: rejected`, refresh the `updated` date, and append the reason under a `## Rejection` section (date + reason) so the history stays visible
4. **Present** the reason and the suggested next step — `plan-revise` to address it, `plan-research` if context was missing, or leave it rejected — then STOP

## Gotchas

- A `rejected` plan fails `plan-subplans-create` (which requires `status: approved`) — it cannot be executed until addressed and re-accepted with `plan-accept`
- Never delete the plan on rejection — the reason and history are the point
