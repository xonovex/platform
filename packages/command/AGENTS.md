# Commands

Claude Code slash-command plugins (`@xonovex/command-*`) for the Xonovex marketplace. Each command plugin delegates to guideline skills via its `.claude-plugin/plugin.json` `dependencies` (see the root `AGENTS.md` Integration Points).

## Version bump

- Command plugins share **one lockstep version** with the skill plugins (`packages/skill/*`) and `.claude-plugin/marketplace.json` (currently `3.x`) — they all move together to the same number. Per package, bump `package.json` `version`, `.claude-plugin/plugin.json` `version`, and `.codex-plugin/plugin.json` `version`; bump `marketplace.json` `metadata.version` to match. Full procedure: `packages/skill/AGENTS.md` (Version bump).
- `private` and git-sourced by the marketplace — no npm/CI publish, so the version on `main` *is* the published state. Run `npm install` afterwards to refresh `package-lock.json`.
