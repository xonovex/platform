# Sources

Each source is one `##` section. Use the **docs** form for a framework/library's official documentation (a single URL); use the **authored** form for a cited book, paper, or article (add a `**Title:**` line with author + work). Author/book/company names live only in this file, never in `SKILL.md` or `references/`.

## {Tech} {Documentation type, e.g. 'Full Documentation' / 'API Reference'}

- **URL:** {https://example.com/llms-full.txt | https://example.com/docs}
- **Last reviewed:** {YYYY-MM-DD}
- **Used for:**
  - `SKILL.md` → {all sections | specific sections}
  - All files under `references/`
- **Aspects extracted:**
  - {Topic 1} → `references/{topic-1}.md`
  - {Topic 2} → `references/{topic-2}.md`
  - {Topic 3} → `references/{topic-3}.md`

## {Concept / source name — for a cited book, paper, or article}

- **Title:** {Author(s) — "Work Title" (publisher / venue, year)}
- **URL:** {https://example.com}
- **Last reviewed:** {YYYY-MM-DD}
- **Used for:**
  - `references/{topic}.md`
  - `SKILL.md` → {section}
- **Aspects extracted:**
  - {the specific claim / pattern this source backs}

## Refresh Workflow

1. Re-fetch the upstream source(s)
2. Diff against the prior pull (or scan for newly added sections)
3. For each changed area, update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
