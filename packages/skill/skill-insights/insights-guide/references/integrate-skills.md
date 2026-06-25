# integrate-skills: Convert Insights into a Guideline Skill

Convert insights from a category into a new or updated guideline skill with progressive-disclosure structure.

## Workflow

1. **Discover** — search `insights/` for category files, extract Problem/Solution/Example, group by topic
2. **Generate Structure**
   - Metadata: `{category}-guide` name and description in skill-guide's format (it owns the metadata shape)
   - SKILL.md: Requirements (optional), Essentials (3-7 items), Examples (code), Progressive disclosure (reference links with load-when triggers)
   - Reference files: Guideline, Rationale, How to Apply, Example (bad vs good), Related
3. **Merge** — if skill exists (unless user asked to overwrite): combine metadata, deduplicate Essentials (keep 3-7), append examples, add reference-file links
4. **Output** — preview shows structure without writing; otherwise creates directory and files

## Structure

Follow skill-guide for the SKILL.md frontmatter/Essentials/Progressive-disclosure structure and the `references/{topic}.md` template — it owns guideline-skill authoring. Map insights onto that structure as follows:

- **Category → skill** — one insights `category` becomes one `{category}-guide` skill
- **Topics → references** — group a category's insights by `topic`; each topic with enough mass becomes a `references/{topic}.md`
- **Problem/Solution/Example → reference body** — the extracted fields fill the Guideline / Rationale / How to Apply / Example slots
- **Essentials** — promote the 3-7 strongest per-category guidelines to the SKILL.md Essentials list

## Example Output

```
[OK] Skill created: <skills-dir>/api/SKILL.md

- 12 insights converted, 4 topics
- Created references/status-codes.md, references/validation-safety.md
```

## Error Handling

- Missing category → ask user
- No insights found → suggest running `extract` for the category
- Output not writable → report error

## Gotchas

- Insights that are one-off corrections don't deserve a whole skill — fold them into AGENTS.md via `integrate-instructions` instead
- Generating a skill with only 1-2 essentials produces noise — wait until the category has enough insights to fill 3-7 essential bullets
- Each reference link must have an explicit load-when trigger; a bare `see references/x.md` defeats progressive disclosure
