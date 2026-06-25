# harness-formats: Prompt File Formats Across Agent Harnesses

Reference for the file formats, locations, and conventions used by different agent harnesses to define reusable **user-invocable** prompts (a.k.a. slash commands, custom commands, prompt files, rules).

**Out of scope:** project-wide always-applied context files (e.g. `AGENTS.md`, `CLAUDE.md`, `.cursorrules`). Those are not invoked by name — for those, see the `skill-instruction` skill.

## Harness Matrix

| Harness            | File format                                                                                                                  | Project location                                                                      | User location                                             | Key features                                                                                                                                                                                                                                                                                 |
| ------------------ | ---------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------- | --------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Claude Code**    | `<name>.md` (legacy) or `<name>/SKILL.md` (current)                                                                          | `.claude/commands/` or `.claude/skills/`                                              | `~/.claude/commands/` or `~/.claude/skills/`              | YAML frontmatter: `description`, `allowed-tools`, `argument-hint`, `model`, `disable-model-invocation`. `$ARGUMENTS` / `$1 $2`. Inline `` !`cmd` `` embedding. Skills merged with commands; if both exist same name, skill wins.                                                             |
| **Cursor**         | `<name>.mdc` (reliable) or `<name>/RULE.md` directory format (documented but flaky — flat `.mdc` still required as of 2.3.x) | `.cursor/rules/`                                                                      | global via Cursor settings                                | YAML frontmatter: `description`, `globs`, `alwaysApply`. 4 activation modes: Always Apply, Auto Attached (globs), Agent Requested, Manual (`@rule-name`). Team Rules > Project > User precedence.                                                                                            |
| **GitHub Copilot** | `<name>.prompt.md`                                                                                                           | `.github/prompts/`                                                                    | `~/.config/Code/User/prompts/`                            | YAML frontmatter: `agent`, `model`, `tools`, `description`. Vars: `${input:name}` / `${input:name:placeholder}`. Tool refs: `#tool:name`. Distinct from repo-wide `.github/copilot-instructions.md`. VS Code, Visual Studio 17.10+, JetBrains.                                               |
| **Gemini CLI**     | `<name>.toml`                                                                                                                | `.gemini/commands/`                                                                   | `~/.gemini/commands/`                                     | TOML keys: required `prompt` (multiline string), optional `description`. `{{args}}` substitution. Inline `!{cmd}` shell execution. Subdirectory → colon namespace: `git/commit.toml` → `/git:commit`. Managed via `/commands` (list / reload). MCP prompts also supported.                   |
| **Continue Dev**   | YAML `prompts:` entries (current) or `customCommands` / `slashCommands` (legacy JSON)                                        | `.continue/config.yaml` or `.continue/config.json`                                    | `~/.continue/config.yaml`                                 | Handlebars `{{{ input }}}`. `invokable: true` makes a markdown file `/`-invokable. MCP-based slash commands experimental. Built-in `slashCommands`: `commit`, `share`, `cmd`, `issue`, `onboard`, `http`.                                                                                    |
| **Aider**          | No custom `/`-commands. Use `--load <file>` of `/commands` on launch; `read:` to preload conventions                         | `.aider.conf.yml` (repo root)                                                         | `~/.aider.conf.yml`                                       | `read: CONVENTIONS.md` to preload context. `alias:` for model shortcuts. No per-name slash commands.                                                                                                                                                                                         |
| **Cline**          | `<name>.md` (kebab-case)                                                                                                     | `.clinerules` (single) or `.clinerules/` (directory); workflows in `workflows/`       | global rules in Cline settings                            | Toggleable popover shows active rules. `/newrule` to create. Falls back to single `.clinerules` if directory absent.                                                                                                                                                                         |
| **Roo Code**       | `<name>.md` / `.txt`                                                                                                         | `.roo/rules/` (recursive, loaded alphabetically) + `.roo/rules-<mode>/` (mode-scoped) | `~/.roo/rules/`                                           | Custom Modes with scoped tool permissions. Falls back to `.roorules` / `.clinerules`.                                                                                                                                                                                                        |
| **OpenCode** (sst) | `<name>.md`                                                                                                                  | `.opencode/commands/`                                                                 | `~/.config/opencode/commands/` or `~/.opencode/commands/` | YAML frontmatter: `description`, `agent`, `model`, `subtask`, `allowed-tools`. `$ARGUMENTS` substitution. `@filename` to inline file contents. `` !`cmd` `` for shell. Subdirectory → colon namespace: `frontend/component.md` → `/project:frontend:component`.                              |
| **Pi** (pi-mono)   | `<name>.md`                                                                                                                  | `.pi/prompts/`                                                                        | `~/.pi/agent/prompts/`                                    | YAML frontmatter: optional `description` (falls back to first non-empty line), `argument-hint` (`<required>` / `[optional]`). Substitution: `$@` for all args, `{{name}}` for named placeholders. Invocation `/name`. Skills separate (`/skill:name`). Disable with `--no-prompt-templates`. |

## Format Axes (where harnesses diverge)

When migrating a prompt across harnesses, these are the dimensions that change:

1. **Body language:** Markdown (most) vs TOML (Gemini)
2. **Metadata mechanism:** YAML frontmatter (most) vs TOML keys (Gemini) vs none (Aider, some Cline)
3. **File extension:** `.md` / `.mdc` / `.prompt.md` / `.toml` / `.txt`
4. **Scoping tiers:** project + user (most) vs project-only (some)
5. **Namespacing:** flat kebab-case (most) vs path-to-colon (Gemini: `git/commit.toml` → `/git:commit`; OpenCode: `frontend/component.md` → `/project:frontend:component`)
6. **Argument syntax:**
   - Claude Code: `$ARGUMENTS`, `$1`, `$2`
   - OpenCode: `$ARGUMENTS`
   - Pi: `$@` (all args), `{{name}}` (named placeholders)
   - Gemini CLI: `{{args}}`
   - GitHub Copilot: `${input:name}`, `${input:name:placeholder}`
   - Continue: `{{{ input }}}`
   - Cursor / Cline / Roo: no built-in substitution (use the body directly)
7. **Inline shell execution:** `` !`cmd` `` (Claude Code, OpenCode) / `!{cmd}` (Gemini) / not supported elsewhere
8. **Inline file references:** `@filename` (OpenCode pulls file contents into the prompt; Pi uses similar pattern for context files) / not supported elsewhere as a substitution
9. **Activation modes:** explicit-only (most) / auto-attached on file globs (Cursor) / always-apply (Cursor `alwaysApply`) / agent-requested (Cursor description-based)
10. **Structure:** single-file (most) vs directory-per-prompt (Claude skills, Cursor `RULE.md` — documented but flaky, flat `.mdc` still required as of 2.3.x, Roo modes)
11. **Reload:** hot-reload (Cursor, Gemini `/commands reload`) vs restart-required (most others)
12. **Frontmatter fields supported:**
    - Common: `description`
    - Permissions: Claude / OpenCode `allowed-tools`, Copilot `tools`
    - Model override: Claude / OpenCode / Copilot `model`
    - Agent / subtask: OpenCode `agent`, `subtask`
    - Activation control: Cursor `globs` / `alwaysApply`, Claude `disable-model-invocation`
    - Args hint: Claude / Pi `argument-hint` (Pi uses `<required>` / `[optional]` syntax)
    - Description fallback: Pi falls back to first non-empty line if `description` omitted

## Migration Notes

- **Claude Code → OpenCode:** near drop-in — same frontmatter shape, same `$ARGUMENTS`, same `` !`cmd` ``; just move from `.claude/commands/` to `.opencode/commands/` and add `@file` references where useful
- **Claude Code → Gemini CLI:** rewrite frontmatter as TOML keys, convert `$ARGUMENTS` → `{{args}}`, change file ext `.md` → `.toml`, embed body as a `prompt = """…"""` string
- **Claude Code → Cursor:** rename `.md` → `.mdc`, replace `allowed-tools` with `globs` + `alwaysApply` if narrowing scope, drop `$ARGUMENTS` (Cursor invokes rules in context, no positional args)
- **Claude Code → Pi:** keep `.md`, convert `$ARGUMENTS` → `$@`, drop `allowed-tools` (Pi has a minimal 4-tool harness), move from `.claude/commands/` to `.pi/prompts/`
- **OpenCode → Pi:** same body language, swap path-to-colon namespacing for flat names, `$ARGUMENTS` → `$@`
- **Cursor `.mdc` → Copilot:** `.mdc` → `.prompt.md`, replace `globs` with explicit file references in body, drop `alwaysApply` (Copilot prompts are explicit-invocation only)
- **Continue legacy JSON → YAML:** move `customCommands` entries into `prompts:` markdown files with `invokable: true`

## Gotchas

- A flat-list prompt from Cursor's "alwaysApply" mode becomes an _explicit_ invocation in Copilot — agents won't auto-pick it up; reword the description to make discovery work
- Gemini's TOML requires triple-quoted multiline `prompt`; trying to dump raw Markdown directly into a `.toml` will break parsing
- Aider has no `/<name>` mechanism — translating a "command" means turning it into a `read:` conventions file or a `--load` script of `/commands`
- Roo Code custom modes scope **tool permissions**, not just prompt content; a Cursor rule with no permission model needs explicit scoping when ported
- "Slash command" and "skill" are increasingly conflated in Claude Code (merged in v2.1.101) — name collision: skill wins over command of the same name
- Don't conflate user-invocable prompts with always-applied project context (`AGENTS.md`, `CLAUDE.md`, `.cursorrules`) — different categories, different skills (see `skill-instruction` for the latter)
