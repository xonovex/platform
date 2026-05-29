---
name: adr-guide
description: "Use when creating or writing an Architecture Decision Record (ADR), recording an architectural decision, or documenting why a design was chosen. Triggers on 'write/create an ADR', 'document this decision', adding files to an `adr/` folder, or capturing a design choice, even when the user doesn't say 'ADR' (e.g. 'record why we chose X'). Works for not-yet-implemented proposals and for existing functionality with no ADR yet (written retroactively, still in proposal voice). Produces a house-style ADR: numbered title, a Status / Date / Decision-makers table, and Context / Proposal / Example uses / Decision sections, in present-tense proposal voice with no em dashes. Domain-agnostic. Skip non-decision docs (runbooks, API references, READMEs)."
---

# ADR authoring

Write Architecture Decision Records in a consistent house style. An ADR captures one decision: the forces behind it, what is proposed, and concrete examples. It reads as a standing proposal regardless of whether the work is built yet.

## Essentials

- **Copy the template** - Start from [assets/adr-template.md](assets/adr-template.md) and fill every section.
- **Number sequentially** - Filename `NNN-kebab-title.md`; the `# NNN: Title` heading matches the number and the padding of existing ADRs in the folder.
- **Status table** - One row with Status, Date (`DD Mon YYYY`), and Decision makers; the `## Decision:` line mirrors Status in brackets.
- **Proposal voice, always** - Write present/future tense as if the change is not yet built, even when documenting shipped code. See [references/style.md](references/style.md).
- **No em dashes** - Use commas, parentheses, colons, or semicolons instead.
- **Standing design vs today's reality** - Context states the standing principles, then marks where today deviates; never blend the target state into the current-state description. See [references/sections.md](references/sections.md).
- **Ground it** - Pin the contract with real schema/code excerpts and an example payload plus a numbered walkthrough under "Example uses".

## Gotchas

- Past tense, even for shipped work, turns an ADR into a changelog; keep it present/future.
- Passive past participles ("the schema is generated") read as completed work; prefer active present ("the build generates the schema").
- "current" and "today" in the same context read redundantly; name the standing design once, mark the deviation once.
- Do not assert the target state as already true inside the current-state description.
- Keep Status and the `## Decision:` line in sync (`Pending`/`[Pending]`, `Accepted`/`[Accepted]`).
- ADRs are immutable once Accepted: record a reversal as a new superseding ADR rather than rewriting one.

## Example

```markdown
# 003: Store all timestamps in UTC

| Status   | Date        | Decision makers     |
| :------- | :---------- | :------------------ |
| Accepted | 12 Jun 2026 | A. Author, B. Owner |

## Context:

Under the data-handling guidelines:

- Storage holds canonical values; presentation layers localize them.

Today, though, some services persist local-time timestamps, against the guideline above.

We need one timestamp convention every service reads and writes, so values cannot be misread across zones.

## Proposal:

This ADR proposes every service stores and transports timestamps in UTC and localizes only at the display edge.

**With the following principles:**

- One canonical representation, localized only for display.

## Decision:

[Accepted]
```

## Progressive disclosure

- [references/style.md](references/style.md) - Load when writing or revising the prose (tense, em dashes, current-vs-today, status lifecycle, numbering, dates)
- [references/sections.md](references/sections.md) - Load when deciding what goes in each section, or writing for existing vs not-yet-built functionality
- [assets/adr-template.md](assets/adr-template.md) - Copy as the starting point for a new ADR
