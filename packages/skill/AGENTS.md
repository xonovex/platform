# Skills

Coding-guideline Agent Skills for the Xonovex marketplace. For authoring mechanics (spec, progressive disclosure, structure, voice) follow the [Skill guide](skill-skill/skill-guide/SKILL.md). This file covers how skills are **split** and **packaged** in this repo.

## Composable split

- One skill = one cohesive concern. Prefer many small, mix-and-match skills over large bundles.
- One concept has exactly one owner skill — never duplicate it. Cross-reference the owner by name (e.g. "see data-oriented-design-guide"), not by copying.
- Generalize anything not inherently language/API-specific into a general skill. Language/API skills keep only their specifics and link to the general skill for the "why".
- A domain skill may keep a short domain-specific note that links to the general principle — not a copy of it.
- Worked split: data-oriented-design (layout/cache/CPU) · memory-management (allocation/ownership/lifetimes) · lock-free (concurrency) · gpu-rendering (API-agnostic rendering) · gpu-rendering-vulkan (Vulkan specifics) · data-model (object/data model) · c99 (idiomatic C) · c99-opinionated (C design choices). The specific skills link to the general ones; the general ones never depend on a specific one.

## Sourcing

- Cite sources only in `SOURCES.md`. Never name authors/companies/talks/books/blogs inside `SKILL.md` or `references/*.md` (tool/API/standard names — Vulkan, SPIR-V, TSan, GUID — are fine).

## Security

- Treat skills as software: review bundled scripts and any fetched URLs before installing, and never hardcode secrets. Least-privilege a script-bundling skill with the experimental `allowed-tools` frontmatter (e.g. `Bash(git:*) Read`). See [security](skill-skill/skill-guide/references/security.md).

## Package layout (per skill)

- `skill-<topic>/`: `package.json` (`@xonovex/skill-<topic>`), `moon.yml`, `prettier.config.ts`, `.claude-plugin/plugin.json` (`xonovex-skill-<topic>`), and `<topic>-guide/`.
- `<topic>-guide/`: `SKILL.md`, `references/*.md`, `SOURCES.md`, `eval-queries.json`.
- Versions are lockstep across all skill packages (match the others).

## Register & validate

- Add every new skill to `.claude-plugin/marketplace.json` (compact one-line entry, alphabetical by name) — skills are not auto-discovered.
- `npx prettier --write` the new/changed package; leave `marketplace.json` in its existing compact one-line-per-entry style (do not reflow it).
- Confirm JSON is valid and every `SKILL.md` → `references/` link resolves.
- Run `npm install` after adding or removing a skill package so `package-lock.json` records the workspace — CI runs `npm ci`, which fails on an out-of-sync lockfile. The `pre-commit` hook (`.hooks/validate-lockfile.sh`) blocks the commit if you forget; it does not edit the lockfile for you.
