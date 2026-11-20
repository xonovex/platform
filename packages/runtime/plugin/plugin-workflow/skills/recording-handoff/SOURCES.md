# Sources

## Open Knowledge Format version 0.2

- **URL:** https://github.com/GoogleCloudPlatform/open-knowledge-format
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `SKILL.md` → Essentials, Gotchas
  - `references/layout.md` → the bundle and its reserved names
  - `references/frontmatter.md` → concept IDs, actors, `verified`, timestamps, unknown keys, sources and footnotes, status
  - `references/generated-files.md` → indexes without frontmatter, grouped entries, the date-grouped log
  - `assets/templates/decision.md` → the `verified` entry
- **Aspects extracted:**
  - A change is a bundle; `index.md` and `log.md` are reserved and hold no concept → layout
  - The concept ID is the path without `.md` → frontmatter, Essentials
  - The `human:` prefix is reserved for content a person wrote or confirmed → frontmatter, Essentials
  - Timestamps carry an explicit UTC offset; unknown keys survive every edit; a superseded document is deprecated, never deleted → frontmatter, Gotchas
  - Indexes group entries under headings; the log is grouped by date, newest first → generated files

## Xonovex change-workflow conventions

- **Provenance:** Repository-original contract of the change directory shared by the workflow steps: names and suffixes, write rules, the entry step, the project settings, the family level, the shared schemas and templates, and the validator
- **Last reviewed:** 2026-11-19
- **References:** references/entry-step.md, references/ids-and-write-rules.md, references/settings.md, references/validator.md, references/layout.md, references/generated-files.md, references/frontmatter.md
- **Aspects extracted:** The entry step, the slug-suffix names, append-only versus edited-in-place files, the skills map with its defaults, the family level, the shared schemas and the three validation layers are conventions of this workflow.
