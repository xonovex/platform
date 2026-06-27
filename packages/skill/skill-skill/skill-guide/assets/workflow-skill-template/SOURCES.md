# Sources

Optional: Include only if the skill distills an external source. House-process workflow skills (a planning lifecycle, a git procedure) that encode no external reference omit this file. When a source IS distilled, use the **docs** form for official documentation and the authored **Title** form for a cited book, paper, or article. Author/book/company names live only in this file.

## {Source name}

- **Title:** {Author(s) — "Work Title" (publisher / venue, year)} {— omit for a pure docs URL}
- **URL:** {https://example.com}
- **Last reviewed:** {YYYY-MM-DD}
- **Used for:**
  - `references/{operation}.md`
  - `SKILL.md` → {section}
- **Aspects extracted:**
  - {the specific claim / pattern this source backs}

## Refresh Workflow

1. Re-fetch the upstream source(s)
2. Diff against the prior pull (or scan for newly added sections)
3. For each changed area, update the corresponding `references/<operation>.md`
4. Bump **Last reviewed** date above
