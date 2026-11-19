# Sources

## Open Knowledge Format version 0.2

- **URL:** https://github.com/GoogleCloudPlatform/open-knowledge-format
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `references/plan-inputs.md` → Front matter
  - `references/plan-sections.md` → Current state
- **Aspects extracted:**
  - Each claim is attributed through a `sources` entry and a footnote keyed to its id → plan-sections

## Xonovex change-workflow conventions

- **Provenance:** Repository-original rules for writing the plan of a change and the first outline of its work, built on the contract of **recording-handoff**: the plan holds only what no other file holds, the generated block of links, settled things stated flat, and the ready or waiting split of the work
- **Last reviewed:** 2026-11-19
- **References:** references/plan-inputs.md, references/plan-sections.md, references/work-outline.md, references/writing-rules.md
- **Aspects extracted:** The plan links decisions and work and never copies them; the generated block is rewritten whole from the files that exist; only the plan-writing and plan-revising steps change `plan.md`, so an existing plan stops the run; a waiting work file names the decisions it waits on in `waits_on`; an unrecorded question is reported and never written as a decision file; the run starts with the entry step of **recording-handoff**.
