# guideline-skills: Creating Coding-Guideline Skills

A guideline skill teaches **coding style / rules** for a language, framework, or paradigm (e.g. typescript, ddd, code-quality). For a **procedure / workflow** skill the agent or a command delegates to (plan, git, …), see [workflow-skills.md](workflow-skills.md) instead. A skill may **combine both** — one that teaches a format _and_ its authoring procedure (as llmstxt does) keeps these guideline sections and adds an `## Operations` section from the workflow template.

An **output-artifact hybrid** — a guideline skill that also prescribes a structured deliverable (e.g. pull-request, adr) — uses this skeleton and adds the artifact's sections after `Example` (e.g. `What / Why / Changes / Testing`), per [instruction-patterns.md](instruction-patterns.md) "Templates for Output Format".

## Template

A ready-to-scaffold template lives under `assets/guideline-skill-template/`:

- [`SKILL.md.template`](../assets/guideline-skill-template/SKILL.md.template) — frontmatter, Essentials, Gotchas, Example, Progressive Disclosure (`## Requirements` is language/framework-only — delete it for general / pattern / process skills)
- [`SOURCES.md`](../assets/guideline-skill-template/SOURCES.md) — source tracking with a **docs-URL** form and an authored **`Title:`** form for cited books/papers/articles
- [`eval-queries.json`](../assets/guideline-skill-template/eval-queries.json) — 12 trigger-eval queries (8 train + 4 validation, mix of should-trigger and near-miss)
- [`references/{topic}.md`](../assets/guideline-skill-template/references/{topic}.md) — reference template (`## sub-headers`, one per facet: statement / rationale / bad→good example; `## Contents` once >200 lines)

To scaffold a new guideline skill: copy the directory, rename `{topic}.md` files, and fill in `{placeholders}`.

## Topic Categories

Group guidelines into one reference file per topic. Common buckets:

- architecture, performance, testing, security, error-handling
- naming, state-management, accessibility
- validation, routing, observability, deployment

## Neutral Examples for General / Pattern Skills

A general or architectural-**pattern** skill (anything not tied to one language or API) must illustrate with a **neutral, widely-recognized domain**, never the codebase or project that motivated it.

- **Pick one neutral running example and reuse it across the skill** — e.g. an order service + a repository, notifications with email/sms/push channels, a storage backend, shapes, documents. A reader who has never seen your repo must be able to follow every snippet.
- **Map the motivating project onto the neutral domain; never name its types or dirs** — if a real refactor inspired the skill, translate it. `Isolator` / `nixprov` / `internal/sandbox` becomes `Repository` / `s3` / `internal/store`; "our agent sandbox" becomes "a data-export tool". The principle is the subject, not the project.
- **No in-repo provenance in the body** — drop phrases like "in this repo", "our service", and any `## Motivating in-repo example` SOURCES entry; they bind the skill to one project. (Language/framework skills are the exception — their specifics _are_ the subject.)
- **Test:** hand the skill to someone who has never seen your codebase. If an example only parses with your project's vocabulary, it is not yet project-agnostic.

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
- `## sub-headers`, one per facet of the topic, each with its own statement, rationale, and bad→good example
- Add a `## Contents` list once the file passes ~200 lines so a partial read shows its full scope
- One topic per file; one level deep under `references/`
- Filename is kebab-case matching the topic

## Style Tips

- Remove source-specific paths and project names — keep content reusable across projects
- Mine corrections / PR comments / fix commits for **Gotchas** entries
- Apply the parent SKILL.md Core Principles (add what the agent lacks, defaults over menus, procedures over declarations)
