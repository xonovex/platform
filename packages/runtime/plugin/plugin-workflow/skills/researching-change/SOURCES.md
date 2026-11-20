# Sources

## How we built our multi-agent research system, Anthropic engineering

- **URL:** https://www.anthropic.com/engineering/multi-agent-research-system
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `SKILL.md` → Essentials, Gotchas
  - `references/explorer-roles.md` → writing a brief, fan-out limits
  - `references/migration.md` → one read-only explorer for each role
  - `references/migration-role-briefs.md` → the bounded brief of each role
  - `assets/templates/explorer-brief.md`, `assets/templates/producer-brief.md`, `assets/templates/target-brief.md`, `assets/templates/consumer-brief.md` → the required output and the boundary
- **Aspects extracted:**
  - A brief states an objective, an output format, the sources to read and a clear task boundary, because vague briefs make parallel explorers duplicate each other's work → explorer roles, briefs, Gotchas
  - A lead agent runs explorers in parallel and scales their number to the question, so the number stays bounded and the lead can read every report → fan-out limits, Essentials

## Open Knowledge Format version 0.2

- **URL:** https://github.com/GoogleCloudPlatform/open-knowledge-format
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `references/evidence.md` → sources and footnotes
  - `assets/templates/research.md` → the `sources` list and the footnotes
- **Aspects extracted:**
  - A body attributes a claim with a footnote whose label is a source id, so the attribution survives reordering → evidence, research template

## Xonovex research conventions

- **Provenance:** Repository-original procedure of the research step and of migration research: the intent check, the evidence rules, the research file and its write rules, and the producer, target and consumer roles with their report sections and failure modes
- **Last reviewed:** 2026-11-19
- **References:** references/intent-check.md, references/evidence.md, references/explorer-roles.md, references/research-file.md, references/migration.md, references/migration-role-briefs.md, references/migration-report-sections.md, references/migration-failure-modes.md
- **Aspects extracted:** The research runs the entry step of recording-handoff first and loads skills from its project settings; a research file is append-only and a correction is a new file that lists the old one in `corrects`; a migration adds no document type and writes a research file with `kind: migration`; observed and inferred facts stay apart; the gap table, the producer changes and the five failure modes are conventions of this workflow.
