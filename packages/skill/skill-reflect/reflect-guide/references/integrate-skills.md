# integrate-skills: Convert Insights into a Guideline Skill

Convert insights from a category into a new or updated guideline skill with progressive-disclosure structure.

## Workflow

1. **Gather** — by DEFAULT, extract lessons from the current session (see [extract.md](extract.md)) and hold them in memory; with `--from-reflections`, read `reflections/*.md`; with `--persist`, also write them to `reflections/`. No `category` → take all session insights. Group by topic.
2. **Generate Structure**
   - Metadata: `{category}-guide` name and description in skill-guide's format (it owns the metadata shape)
   - SKILL.md: Requirements (optional), Essentials (3-7 items), Examples (code), Progressive disclosure (reference links with load-when triggers)
   - Reference files: Guideline, Rationale, How to Apply, Example (bad vs good), Related
3. **Merge** — if skill exists (unless user asked to overwrite): combine metadata, deduplicate Essentials (keep 3-7), append examples, add reference-file links
4. **Output** — preview shows structure without writing; otherwise creates directory and files

## Structure

Follow skill-guide for the SKILL.md frontmatter/Essentials/Progressive-disclosure structure and the `references/{topic}.md` template — it owns guideline-skill authoring. Map insights onto that structure as follows:

- **Owning skill first** — apply each insight to the EXISTING skill that owns its domain (merge into its `SKILL.md` Essentials/Gotchas and the relevant `references/{topic}.md`), even across categories. Create a NEW `{category}-guide` only when a category has 3-7 worth-keeping insights AND no owner skill exists. A general lesson belongs in its general owner, not a project-only file (this repo's composable-split rule).
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

- Missing category → take all session insights, routed to their owning skills
- No insights found in the session → report no lessons detected; with `--from-reflections`, suggest running `extract` first
- Output not writable → report error

## Gotchas

- Insights that are one-off corrections don't deserve a whole skill — fold them into AGENTS.md via `integrate-instructions` instead
- Generating a skill with only 1-2 essentials produces noise — wait until the category has enough insights to fill 3-7 essential bullets
- Each reference link must have an explicit load-when trigger; a bare `see references/x.md` defeats progressive disclosure
- Apply directly by default — `reflections/*.md` is written only with `--persist` or after an explicit `extract`; don't force a store-then-integrate two-step
- Publishing a change to a marketplace skill needs the lockstep version bump across the skill + command plugins and `marketplace.json` (see the skills package `AGENTS.md`) — flag it; don't silently leave the source ahead of the published version
