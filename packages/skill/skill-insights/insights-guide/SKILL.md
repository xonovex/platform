---
name: insights-guide
description: "Use when reflecting on a session or distilling lessons into reusable form. Triggers on prompts about extracting insights, capturing mistakes / discoveries / patterns, running a post-mortem, post-session retrospectives, or folding lessons into AGENTS.md or into a guideline skill, even when the user doesn't say 'insights'."
---

# Insights Guidelines

Extract development lessons from a session and integrate them either as AGENTS.md bullets (project instructions) or as a guideline skill (reusable across sessions).

## Core Principles

- **Extract from experience** — analyze the session for mistakes, discoveries, and corrections, not hypotheticals
- **Identify general patterns** — capture only insights that apply beyond the current task
- **Structured storage** — save each insight as a file with `category` / `topic` / `applies_to` frontmatter
- **Two integration paths** — small one-off lessons → AGENTS.md (project instructions); recurring patterns with enough mass → guideline skill
- **`applied: false` is the pending flag** — flip it only after the insight is actually integrated

## Gotchas

- Catching a "mistake" that was correct in context produces noise — only capture corrections you'd want a future agent to remember
- `applies_to: ["general"]` makes integration impossible — be specific with routing keys
- One-off corrections don't deserve a whole skill — fold them into AGENTS.md instead
- A skill with only 1-2 essentials is filler — wait until the category has 3-7 worth-keeping bullets
- A generated skill must follow skill-guide's naming and metadata conventions (`{category}-guide`, its description shape), not an ad-hoc inline format, to avoid drift from sibling skills

## Operations

- **Extract** — capture development lessons from the current session — see [references/extract.md](references/extract.md)
- **Integrate-instructions** — fold insights into AGENTS.md as bullet points — see [references/integrate-instructions.md](references/integrate-instructions.md)
- **Integrate-skills** — convert insights into a new or updated guideline skill — see [references/integrate-skills.md](references/integrate-skills.md)

## Progressive Disclosure

- Read [references/extract.md](references/extract.md) - Load when analyzing the current session for development lessons and saving them with frontmatter
- Read [references/integrate-instructions.md](references/integrate-instructions.md) - Load when folding extracted insights from a category into AGENTS.md as bullet points
- Read [references/integrate-skills.md](references/integrate-skills.md) - Load when converting extracted insights from a category into a new or updated guideline skill
