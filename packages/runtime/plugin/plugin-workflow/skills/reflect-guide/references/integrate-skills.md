# integrate-skills: Convert Insights into a Guideline Skill

Convert a category's insights into a new or updated guideline skill.

## Workflow

1. **Gather**: by DEFAULT extract from the current session and hold in memory; `--from-reflections` reads `reflections/*.md`; `--persist` also writes them there. No `category` → all session insights. Group by topic.
2. **Route**: apply each insight to the EXISTING skill that owns its domain (merge into its `SKILL.md` Essentials/Gotchas and the relevant `references/{topic}.md`), even across categories. Create a NEW `{category}-guide` only when a category has 3-7 worth-keeping insights AND no owner skill exists. A general lesson belongs in its general owner, not a project-only file.
3. **Structure**: follow skill-guide (it owns the SKILL.md frontmatter/Essentials/Progressive-disclosure structure and the `references/{topic}.md` template). One `category` → one `{category}-guide`; each `topic` with enough mass → one `references/{topic}.md`; promote the 3-7 strongest guidelines to Essentials.
4. **Merge**: if the skill exists, combine metadata, dedupe Essentials (keep 3-7), append examples, add reference links.

## Gotchas

- One-off corrections don't deserve a whole skill: fold them into AGENTS.md via `integrate-instructions` instead
- A skill with only 1-2 essentials is filler: wait for 3-7 worth-keeping bullets
- Each reference link needs an explicit load-when trigger; a bare `see references/x.md` defeats progressive disclosure
- Publishing a change to a marketplace skill needs the lockstep version bump across the skill + command plugins and `marketplace.json` (see the skills package `AGENTS.md`): flag it; don't leave source ahead of published
