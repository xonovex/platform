# style: ADR writing rules

## Present-tense proposal voice

- **Guideline**: Write the whole ADR in present (and where needed future) tense, as a standing proposal. Never past tense, even for functionality that already exists.
- **Rationale**: An ADR records a decision that stands; past tense turns it into a changelog and dates badly. A retroactive ADR should read identically to a brand-new one.
- **How to apply**: Describe the design as what the system does / will do, not what someone did. Convert passive past participles into active present.
- Bad: "We added a retry policy and the gateway handled failures locally."
- Good: "The gateway retries failed calls and reports the outcome."
- Bad: "The schema is generated from the spec." (passive, reads as done)
- Good: "The build generates the schema from the spec."

## No em dashes

- **Guideline**: Never use em dashes (the long dash). Use commas, parentheses, colons, or semicolons.
- Bad: "The job runs nightly — after the backup completes."
- Good: "The job runs nightly, after the backup completes."

## Standing design vs today's reality

- **Guideline**: In Context, state the standing architecture/principles once, then mark where today's implementation deviates. Keep the two separate.
- **Rationale**: The gap between principle and reality is the motivation for the decision; blending them hides it.
- **How to apply**: Name the standing design plainly ("Under the X architecture:"), reserve a single contrast marker ("Today, though,") for the deviation, and do not assert the target state as if already true inside the current-state description.
- **Counter-example**: Tagging the standing design "current" and then also writing "today" for the same now reads redundantly; pick one. Inserting "This means the system already does X" between the principles and the "today" gap contradicts the gap.

## Status lifecycle

- `Pending` while awaiting decision-maker sign-off; `Accepted` once decided. Keep the `## Decision:` line in sync (`[Pending]` / `[Accepted]`).
- Other values as needed: `Rejected`, `Superseded by NNN`, `Deprecated`.
- ADRs are immutable once Accepted: record a reversal or change as a new ADR that supersedes the old one, rather than rewriting history.

## Numbering, filename, date

- One sequential number per ADR. Filename `NNN-kebab-title.md` (e.g. `003-store-timestamps-in-utc.md`); the `# NNN: Title` heading matches.
- Match the zero-padding of ADRs already in the folder (commonly three digits).
- Date format `DD Mon YYYY` (e.g. `28 May 2026`). Align the table columns for readability.
