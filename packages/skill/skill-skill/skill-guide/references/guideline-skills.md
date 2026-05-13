# guideline-skills: Creating Coding-Guideline Skills

## Template

A ready-to-scaffold template lives under `assets/guideline-skill-template/`:

- [`assets/guideline-skill-template/SKILL.md`](../assets/guideline-skill-template/SKILL.md) — umbrella with frontmatter, Essentials, Gotchas, Example, Progressive Disclosure
- [`assets/guideline-skill-template/SOURCES.md`](../assets/guideline-skill-template/SOURCES.md) — upstream source tracking
- [`assets/guideline-skill-template/eval-queries.json`](../assets/guideline-skill-template/eval-queries.json) — 12 sample trigger-eval queries (8 train + 4 validation, mix of should-trigger and near-miss)
- [`assets/guideline-skill-template/references/{topic}.md`](../assets/guideline-skill-template/references/{topic}.md) — reference file template (Guideline / Rationale / How to Apply / Example / Counter-Example / Related)

To scaffold a new guideline skill: copy the directory, rename `{topic}.md` files, and fill in `{placeholders}`.

## Topic Categories

Group guidelines into one reference file per topic. Common buckets:

- architecture, performance, testing, security, error-handling
- naming, state-management, accessibility
- validation, routing, observability, deployment

## Source Parsing Patterns

When extracting from external docs:

- Headings → topic groups (one reference file per group)
- Code blocks with language markers → examples
- "do / don't", "good / bad", "prefer / avoid" → rule + counter-example pairs in the reference file
- Bullet lists → individual guidelines

## SKILL.md Conventions (matches the template)

- **Essentials bullet format:** `- **Rule** - Brief 5-10 word how-to, see [references/{topic}.md](references/{topic}.md)`
- One rule per bullet; counter-examples go in the reference file, not SKILL.md
- The **Example** section shows one representative idiomatic snippet (the _default_ approach, not a comprehensive demo)
- The **Progressive Disclosure** section is where each reference link's `Load when…` trigger lives — not in the reference file itself

## Reference File Conventions (matches the template)

- Title: `# {topic}: {Title}`
- Body fields: **Guideline**, **Rationale**, **How to Apply**, **Example** (bad / good pair), optional **Counter-Example**, optional **Related**
- One topic per file; one level deep under `references/`
- Filename is kebab-case matching the topic

## Style Tips

- Remove source-specific paths and project names — keep content reusable across projects
- Mine corrections / PR comments / fix commits for **Gotchas** entries
- Procedures over declarations: teach the approach, not the one-off answer
- Defaults over menus: pick one recommended approach in Essentials; alternatives go in the reference file
- Add only what the agent lacks — omit anything inferable from official docs alone
