# Sources

## Claude Code: Slash Commands

- **URL:** https://code.claude.com/docs/en/slash-commands
- **Last reviewed:** 2026-05-13
- **Used for:** `references/harness-formats.md` → Claude Code matrix row, Argument syntax axis, Frontmatter fields axis, Migration Notes
- **Aspects extracted:**
  - `.claude/commands/<name>.md` (legacy) and `.claude/skills/<name>/SKILL.md` (recommended) locations
  - YAML frontmatter fields: `description`, `allowed-tools`, `argument-hint`, `model`, `disable-model-invocation`
  - `$ARGUMENTS`, `$1`, `$2` argument substitution
  - `` !`cmd` `` inline shell execution
  - Skill / command merger in v2.1.101 — skill wins on name collision

## Cursor: Rules

- **URL:** https://cursor.com/docs/rules
- **Last reviewed:** 2026-05-13
- **Used for:** `references/harness-formats.md` → Cursor matrix row, Activation modes axis, Structure axis, Reload axis
- **Aspects extracted:**
  - `.cursor/rules/<name>.mdc` and v2.2+ `<name>/RULE.md` directory format
  - Frontmatter fields: `description`, `globs`, `alwaysApply`
  - Four activation modes (Always Apply, Auto Attached, Agent Requested, Manual)
  - Team / Project / User precedence hierarchy
  - Hot-reload behavior

## GitHub Copilot: Prompt Files

- **URL:** https://docs.github.com/en/copilot/tutorials/customization-library/prompt-files
- **URL:** https://code.visualstudio.com/docs/copilot/customization/prompt-files
- **Last reviewed:** 2026-05-13
- **Used for:** `references/harness-formats.md` → Copilot matrix row, Argument syntax axis, Frontmatter fields axis
- **Aspects extracted:**
  - `.github/prompts/<name>.prompt.md` (workspace) and user-scope location (`~/.config/Code/User/prompts/`)
  - YAML frontmatter: `agent`, `model`, `tools`, `description`
  - `${input:name}` and `${input:name:placeholder}` variable syntax
  - `#tool:name` tool references in body
  - Distinct from `.github/copilot-instructions.md` (repo-wide always-applied)
  - IDE support: VS Code, Visual Studio 17.10+, JetBrains

## Gemini CLI: Custom Slash Commands

- **URL:** https://cloud.google.com/blog/topics/developers-practitioners/gemini-cli-custom-slash-commands
- **URL:** https://geminicli.com/docs/cli/custom-commands/
- **Last reviewed:** 2026-05-13
- **Used for:** `references/harness-formats.md` → Gemini CLI matrix row, Body language axis, Namespacing axis, Inline shell axis, Migration Notes
- **Aspects extracted:**
  - `.gemini/commands/<name>.toml` (project) and `~/.gemini/commands/` (user) locations
  - **TOML schema** (outlier — all other harnesses use Markdown + YAML frontmatter): required `prompt` (multiline triple-quoted string), optional `description`
  - `{{args}}` substitution
  - `!{cmd}` inline shell execution
  - Subdirectory → colon namespace (`git/commit.toml` → `/git:commit`)
  - Built-in `/commands` (list / reload) management

## Continue Dev: Slash Commands

- **URL:** https://docs.continue.dev/customize/slash-commands
- **URL:** https://docs.continue.dev/customize/deep-dives/prompts
- **URL:** https://docs.continue.dev/reference/yaml-migration
- **Last reviewed:** 2026-05-13
- **Used for:** `references/harness-formats.md` → Continue Dev matrix row, Argument syntax axis, Migration Notes
- **Aspects extracted:**
  - Legacy `config.json` `slashCommands` (built-in commit/share/cmd/issue/onboard/http) and `customCommands`
  - Current `config.yaml` `prompts:` field with `invokable: true`
  - Handlebars `{{{ input }}}` substitution
  - MCP-based slash commands (experimental)
  - YAML migration path from legacy JSON

## Aider: YAML Config

- **URL:** https://aider.chat/docs/config/aider_conf.html
- **URL:** https://aider.chat/docs/usage/conventions.html
- **Last reviewed:** 2026-05-13
- **Used for:** `references/harness-formats.md` → Aider matrix row
- **Aspects extracted:**
  - `.aider.conf.yml` location (home or repo root)
  - `read:` field to preload conventions / context files
  - `--load <file>` to execute a file of `/commands` on launch
  - Model `alias:` shortcuts
  - Aider does NOT support user-defined `/<name>` slash commands — confirmed gap

## Cline: Rules / Workflows

- **URL:** https://github.com/cline/clinerules
- **Last reviewed:** 2026-05-13
- **Used for:** `references/harness-formats.md` → Cline matrix row
- **Aspects extracted:**
  - `.clinerules` (single file fallback) and `.clinerules/` (directory) layouts
  - `workflows/` directory for workflows
  - Kebab-case filenames
  - `/newrule` creation command
  - Toggleable rules popover (v3.13+)

## Roo Code: Custom Instructions

- **URL:** https://docs.roocode.com/features/custom-instructions
- **Last reviewed:** 2026-05-13
- **Used for:** `references/harness-formats.md` → Roo Code matrix row
- **Aspects extracted:**
  - `.roo/rules/` recursive directory (alphabetical load order)
  - `.roo/rules-<mode>/` mode-scoped overrides
  - Custom Modes with scoped tool permissions
  - Fallback to `.roorules` / `.clinerules`

## OpenCode (sst): Commands

- **URL:** https://opencode.ai/docs/commands/
- **Last reviewed:** 2026-05-13
- **Used for:** `references/harness-formats.md` → OpenCode matrix row, Argument syntax axis, Inline file references axis, Namespacing axis, Migration Notes
- **Aspects extracted:**
  - `.opencode/commands/<name>.md` (project) and `~/.config/opencode/commands/` (user) locations
  - YAML frontmatter: `description`, `agent`, `model`, `subtask`, `allowed-tools`
  - `$ARGUMENTS` substitution (mirrors Claude Code)
  - `@filename` for inlining file contents
  - `` !`cmd` `` for inline shell
  - Subdirectory → colon namespace (`frontend/component.md` → `/project:frontend:component`)

## Pi (pi-mono): Prompt Templates

- **URL:** https://github.com/badlogic/pi-mono/blob/main/packages/coding-agent/docs/prompt-templates.md
- **URL:** https://pi.dev/docs/latest/prompt-templates
- **Last reviewed:** 2026-05-13
- **Used for:** `references/harness-formats.md` → Pi matrix row, Argument syntax axis, Frontmatter fields axis, Migration Notes
- **Aspects extracted:**
  - `.pi/prompts/<name>.md` (project) and `~/.pi/agent/prompts/` (user) locations
  - YAML frontmatter: optional `description` (falls back to first non-empty line), `argument-hint` with `<required>` / `[optional]` syntax
  - `$@` (all args) and `{{name}}` (named placeholders) substitution
  - `/name` invocation; skills separate via `/skill:name`
  - `--no-prompt-templates` disable flag
