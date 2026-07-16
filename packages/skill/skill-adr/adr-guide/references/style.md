# style: ADR writing rules

## Present-tense proposal voice

Write the whole ADR in present (or future) tense as a standing proposal, never past tense, even for functionality that already ships. Convert passive past participles into active present.

- Bad: "The schema is generated from the spec." (passive, reads as done)
- Good: "The build generates the schema from the spec."

## No em dashes

Use commas, parentheses, colons, or semicolons instead.

- Bad: "The job runs nightly — after the backup completes."
- Good: "The job runs nightly, after the backup completes."

## Standing design vs today's reality

In Context, name the standing architecture/principles once ("Under the X architecture:"), reserve a single contrast marker ("Today, though,") for the deviation, and never assert the target state as already true inside the current-state description. Tagging the standing design "current" and also writing "today" for the same now reads redundantly; pick one.

## Status lifecycle

- `Pending` while awaiting sign-off; `Accepted` once decided. Keep the `## Decision:` line in sync (`[Pending]` / `[Accepted]`).
- Other values as needed: `Rejected`, `Superseded by NNN`, `Deprecated`.
- ADRs are immutable once Accepted: record a reversal as a new superseding ADR, never rewrite history.

## Numbering, filename, date

- Filename `NNN-kebab-title.md` (e.g. `003-store-timestamps-in-utc.md`); the `# NNN: Title` heading matches.
- Match the zero-padding of ADRs already in the folder (commonly three digits).
- Date format `DD Mon YYYY` (e.g. `28 May 2026`).
