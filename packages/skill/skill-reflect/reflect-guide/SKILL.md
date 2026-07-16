---
name: reflect-guide
description: "Use when reflecting on a session or distilling lessons into reusable form. Triggers on prompts about reflecting on a session, extracting insights, capturing mistakes / discoveries / patterns, running a post-mortem, post-session retrospectives, or folding lessons into AGENTS.md or into a guideline skill, even when the user doesn't say 'reflect'."
---

# Reflection Guidelines

Reflect on a session to capture development lessons (insights) and integrate them as AGENTS.md bullets (project instructions) or a guideline skill (reusable across sessions).

## Core Principles

- **Extract from the session** — analyze actual mistakes, discoveries, and corrections, not hypotheticals; keep only patterns that apply beyond the current task
- **Apply directly by default** — `integrate-instructions` / `integrate-skills` extract and write STRAIGHT to AGENTS.md / the owning skill; persisting `reflections/*.md` first is the optional `--persist` audit step, not a prerequisite. Use `extract` only for the stored files, `--from-reflections` to integrate already-stored ones
- **Route by `applies_to`** — each insight's `category` / `topic` / `applies_to` picks the nearest AGENTS.md or the owning skill
- **Two paths** — small one-off lessons → AGENTS.md; recurring patterns with 3-7 worth-keeping bullets → guideline skill
- **`applied: false`** — the pending flag, persisted insights only; flip it after integration
- **Govern promotion by impact** — ordinary reviewable file edits stay direct; managed or executable targets require owner review, versioning, validation, canary where applicable, evidence, and rollback, see [references/governed-promotion.md](references/governed-promotion.md)

## Gotchas

- Catching a "mistake" that was correct in context produces noise — only capture corrections you'd want a future agent to remember
- `applies_to: ["general"]` makes integration impossible — be specific with routing keys
- One-off corrections don't deserve a whole skill — fold them into AGENTS.md instead
- A skill with only 1-2 essentials is filler — wait until the category has 3-7 worth-keeping bullets
- A generated skill must follow skill-guide's naming and metadata conventions (`{category}-guide`, its description shape), not an ad-hoc inline format, to avoid drift from sibling skills
- Direct integration changes a working file; it does not authorize publication, installation, enablement, enforcement, or organization-wide promotion.

## Progressive Disclosure

- Read [references/extract.md](references/extract.md) - Load when analyzing the current session for development lessons and saving them with frontmatter
- Read [references/integrate-instructions.md](references/integrate-instructions.md) - Load when folding extracted insights from a category into AGENTS.md as bullet points
- Read [references/integrate-skills.md](references/integrate-skills.md) - Load when converting extracted insights from a category into a new or updated guideline skill
- Read [references/governed-promotion.md](references/governed-promotion.md) - Load when lessons come from lifecycle, onboarding, policy denials, incidents, exceptions, drift, rollback, support, or module failures, or when the target is managed, organization-wide, executable, enforcing, configuration-changing, or privileged
