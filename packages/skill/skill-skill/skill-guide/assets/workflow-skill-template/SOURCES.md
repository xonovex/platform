# Sources

Keep this file for every catalog skill. Use the external-source form when the workflow distills published material; otherwise replace it with the repository-original provenance form. Author/book/company names live only in this file.

## {Source name}

- **Title:** {Author(s), "Work Title" (publisher / venue, year)} {omit for a pure docs URL}
- **URL:** {https://example.com}
- For a versioned surface, add `Version` plus either `Content SHA256` for web-only documentation or `Checkout` + `Commit` + `Watch` for a repository source.
- **Last reviewed:** {YYYY-MM-DD}
- **Used for:**
  - `references/{operation}.md`
  - `SKILL.md` → {section}
- **Aspects extracted:**
  - {the specific claim / pattern this source backs}

## {Repository-original workflow name: use instead of the source block above}

- **Provenance:** {Repository-original procedure distilled from maintained project practice}
- **References:** all
- **Last reviewed:** {YYYY-MM-DD}

## Refresh Workflow

1. Re-fetch the upstream source(s) and compare the declared content digest or repository drift fields
2. Review every changed watched path or fetched-content digest
3. For each changed area, update the corresponding `references/<operation>.md`
4. Bump **Last reviewed** date above
