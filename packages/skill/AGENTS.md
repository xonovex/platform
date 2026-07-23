# Skills

- Use the [Skill guide](skill-skill/skill-guide/SKILL.md) for authoring mechanics; this file defines repository split and packaging rules.
- Keep one cohesive concern per skill. Each concept has one owner; cross-reference the owner by skill name instead of copying content.
- Move language/API-independent guidance into a general skill. Specific skills keep only their specialization and may depend on the general skill; general skills never depend on a specific one.
- Cite sources only in `SOURCES.md`. Do not name authors, companies, talks, books, or blogs in `SKILL.md` or `references/*.md`; tool, API, and standard names remain allowed.
- Treat skills as software: review bundled scripts and fetched URLs, never hardcode secrets, and restrict script-bundling skills with least-privilege experimental `allowed-tools` frontmatter such as `Bash(git:*) Read`.
- Point each Codex manifest's string `skills` value directly at `"./<topic>-guide"`; `"./"` does not reliably expose nested guide skills.
- Keep versions lockstep across every skill plugin, command plugin, and `marketplace.json`.
- Register every new skill alphabetically in both marketplace files; skills are not auto-discovered.
- Give every guide exactly one reviewed lifecycle and functional-role classification in `composition-catalog.json`; never use a mixed classification.
- Keep exact install-time dependencies in both plugin manifests. Declare semantic provisions and requirements only in `composition-catalog.json`, and only when their contract and compatibility are understood.
- Keep skill implementation versions in package/plugin manifests, semantic contract versions in catalog provisions, and content provenance in each guide's `SOURCES.md`; never copy one version or provenance source into another field.
- Format changed packages and `marketplace.json` with `npx prettier --write`, validate JSON, and resolve every `SKILL.md` -> `references/` link.
- Run `npm install` after adding or removing a skill package so `package-lock.json` matches the workspaces. CI uses `npm ci`, and `.hooks/validate-lockfile.sh` blocks but does not repair stale locks.
