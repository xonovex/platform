# Sources

## Documenting Architecture Decisions

- **URL:** https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `SKILL.md` → Essentials, Gotchas
  - `references/applying.md` → decision files
- **Aspects extracted:**
  - A reversed decision is kept and marked as superseded, so a changed decision is a new file and the old one gets only its deprecation and a pointer to its successor → applying, Gotchas

## Open Knowledge Format version 0.2

- **URL:** https://github.com/GoogleCloudPlatform/open-knowledge-format
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `references/propagation.md` → regenerate
- **Aspects extracted:**
  - Index files and log files are generated, so the run regenerates them by script after its edits and never edits them by hand → propagation

## Xonovex plan-revision conventions

- **Provenance:** Repository-original procedure for folding `FIXME` notes, critique findings and research corrections into the plan and the work files of a change
- **Last reviewed:** 2026-11-19
- **References:** references/inputs.md, references/dispositions.md, references/applying.md, references/work-and-plan.md, references/propagation.md
- **Aspects extracted:** The three inputs and how each is marked as processed, the four dispositions with a rationale, the operator's answer that settles a decision and the evidence that only reopens one, the order of work, the readiness rules of work files, the plan edited outside its generated block, and the dispositions table in the log entry are conventions of this workflow. Decision files follow the shared template and write rules of recording-handoff.
