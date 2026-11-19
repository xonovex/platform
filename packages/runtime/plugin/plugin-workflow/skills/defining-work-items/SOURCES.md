# Sources

## INVEST in Good Stories, and SMART Tasks

- **URL:** https://xp123.com/invest-in-good-stories-and-smart-tasks/
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `SKILL.md` → Essentials
  - `references/writing-criteria.md` → what a verification is
  - `references/cut-rules.md` → group by outcome and discipline
- **Aspects extracted:**
  - A good story is testable, so every criterion of a work item says how it is checked → Essentials, writing criteria
  - A story is valuable, independent and testable, so items are cut by outcome and not by layer, because a layer item cannot be verified alone → cut rules

## Xonovex work-item definition conventions

- **Provenance:** Repository-original procedure for turning the draft work outline of a change into work items and an optional work package, on the shared work item schema, template and write rules of recording-handoff
- **Last reviewed:** 2026-11-19
- **References:** references/cut-rules.md, references/outline-rows.md, references/writing-criteria.md, references/dependencies-and-package.md, references/existing-package.md
- **Aspects extracted:** The cut proposal before any write, the operator's team as the scope, waiting rows kept at their outline with a non-empty `waits_on`, one repository for each item, real blockers only in `depends_on` and an acyclic graph, the optional work package without criteria, attaching items to an existing package, work files edited in place and never renamed, and the separate confirmation that sets `stable` with a `human:` actor are conventions of this workflow.
