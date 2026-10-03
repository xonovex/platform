# Plugins

- Use the [Skill guide](plugin-agentic/skills/skill-guide/SKILL.md) for authoring mechanics; this file defines repository split and packaging rules.
- Keep one cohesive concern per skill. Each concept has one owner; cross-reference the owner by skill name instead of copying content.
- Move language/API-independent guidance into a general skill. Specific skills keep only their specialization and may depend on the general skill; general skills never depend on a specific one.
- Cite sources only in `SOURCES.md`. Do not name authors, companies, talks, books, or blogs in `SKILL.md` or `references/*.md`; tool, API, and standard names remain allowed.
- Keep instruction content in Markdown. Limit JSON to required manifests, package metadata, and evals. Never add TypeScript or MJS implementation files to a skill.
- Treat skills as software: review bundled scripts and fetched URLs, never hardcode secrets, and restrict script-bundling skills with least-privilege experimental `allowed-tools` frontmatter such as `Bash(git:*) Read`.
- Each plugin keeps its skills under `skills/` and its commands under `commands/`. Both harness manifests use `"skills": "./skills/"`.
- Keep versions lockstep across every plugin and both marketplaces, and keep exact install-time dependencies in both plugin manifests.
- Register every plugin alphabetically in both marketplace files. A skill belongs to one plugin, and the harness discovers it under `skills/`.
- Select optional skills at runtime by installed names and routing descriptions.
- Format changed packages and `marketplace.json` with `npx prettier --write`, validate JSON, and resolve every `SKILL.md` -> `references/` link.
- Run `npm install` after adding or removing a plugin package so `package-lock.json` matches the workspaces. CI uses `npm ci`, and `.hooks/validate-lockfile.sh` blocks but does not repair stale locks.
