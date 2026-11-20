# The work outline

## Contents

- What the outline is for
- Choosing the rows
- Teams
- Readiness
- Dependencies
- The fields of a row
- An example row
- Answering what can start now

## What the outline is for

The outline answers one question before the work items are defined: what can start now, and what waits on which decision. It is the first draft of the work files, written as `status: draft` and completed later by **defining-work-items**, which adds the full criteria and the work packages. The outline invents none of those.

## Choosing the rows

One row is one piece of work that one team delivers and verifies independently. Derive the rows from the objective, the scope and the settled and open decisions:

- A settled decision usually yields work that can start.
- An open decision yields work that waits when the work changes with the answer. Work that holds under every option does not wait.
- Split a row when two teams would deliver it, or when one part can start and the other must wait.
- Write every row as `kind: work-item`. Use `kind: work-package` only when the scope already shows a named group of items, and then set `parent` on its items. A package carries an objective and no criteria.

Name the file after the outcome, not the activity: `publish-invoice-events`, not `implement-publisher`.

## Teams

Every row names the team that owns it in `team`. Take the team from the decision owners, the research and the request. When a row has no owner, ask one question that offers the likeliest team. Never leave `team` empty, and never invent a team name to fill the field.

## Readiness

`readiness` reflects decisions only:

- `ready` when no decision the work depends on is `open`, `paused` or `partly-settled`.
- `waiting` when at least one is. List the concept IDs of every such decision in `waits_on`. The validator rejects a waiting file with an empty `waits_on`.
- List the settled decisions that shape the work in `decisions`. They are referenced, never copied.

A paused decision names a fact to wait for; the work waits on the decision, and the file does not repeat the fact. When a decision later settles, **revising-plan** changes the readiness of the files that wait on it.

## Dependencies

`depends_on` lists the work files that must finish first, by concept ID. It orders work and never expresses a decision. A row that depends on a waiting row can itself be `ready`, because it has no open decision of its own. A dependency cycle is a validation error, so draw the order before writing: choose every file name first, then fill `depends_on` with those names.

## The fields of a row

| Key | Value |
| --- | --- |
| `generated` | `{by: writing-plan/<version>, at: <now with offset>}` |
| `status` | `draft`, with no `verified` entry |
| `kind` | `work-item`, or `work-package` as above |
| `team` | The owner team |
| `repository` | A name from the project's repository map; include only when the work touches one repository |
| `readiness` | `ready` or `waiting` |
| `waits_on` | Required when waiting |
| `depends_on` | Work that must finish first |
| `decisions` | Settled decisions that apply |
| `criteria` | At least one `{ id, text, verification }` for a work item; the text is a condition a check can fail, and the verification says the check |
| `skills_to_consult` | Include when the work needs a skill the plan does not already list |

Body: `# Objective`, one sentence that states the outcome. Leave out `# Notes`; **defining-work-items** adds it when the implementer needs context that the criteria do not carry.

## An example row

The change replaces a nightly invoice export with an event feed. The delivery method is settled, and the retention of the old file is open.

```markdown
---
type: Workflow Work Item
schema: workflow/work-item/1
title: Publish invoices as events
description: The billing service publishes each issued invoice to the feed.
generated: {by: writing-plan/0.1.0, at: 2026-10-27T09:30:00+01:00}
status: draft
kind: work-item
team: Billing
repository: billing-service
readiness: ready
decisions:
  - decisions/delivery-method-b3ne
criteria:
  - id: c1
    text: An invoice issued in the test environment appears on the feed within 60 seconds.
    verification: Issue a test invoice and read the feed.
---

# Objective

The billing service publishes every issued invoice to the feed.
```

A second row, `retire-nightly-export`, is `readiness: waiting` with `waits_on: [decisions/nightly-retention-7k2f]` and `depends_on: [work/publish-invoice-events-m2qd]`.

## Answering what can start now

Close the outline with two lists in the reply, from the files you wrote:

- **Can start now:** `ready` rows whose `depends_on` rows are done or absent, with the team of each.
- **Waits:** each `waiting` row with the decisions it waits on, grouped by decision, so an owner sees what their answer unblocks.

A ready row that depends on unfinished work goes in a third line, "starts after", with the row it follows.
