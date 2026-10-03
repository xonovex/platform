# Agentic Plugin

Configure coding-agent harnesses and author their skills, commands, and project instructions.

Install `xonovex-agentic@xonovex-marketplace` in Claude Code or Codex.

## Skills

- [claude-code-guide](skills/claude-code-guide/SKILL.md)
- [codex-guide](skills/codex-guide/SKILL.md)
- [command-guide](skills/command-guide/SKILL.md)
- [copilot-guide](skills/copilot-guide/SKILL.md)
- [instruction-guide](skills/instruction-guide/SKILL.md)
- [kiro-guide](skills/kiro-guide/SKILL.md)
- [llmstxt-guide](skills/llmstxt-guide/SKILL.md)
- [opencode-guide](skills/opencode-guide/SKILL.md)
- [pi-guide](skills/pi-guide/SKILL.md)
- [skill-guide](skills/skill-guide/SKILL.md)

## Commands

| Command                    | Purpose                                                                       |
| -------------------------- | ----------------------------------------------------------------------------- |
| `instructions-assimilate`  | [Augment Project Instructions](commands/instructions-assimilate.md)           |
| `instructions-consolidate` | [Consolidate project instruction files](commands/instructions-consolidate.md) |
| `instructions-init`        | [Create AGENTS.md](commands/instructions-init.md)                             |
| `instructions-simplify`    | [Simplify project instruction files](commands/instructions-simplify.md)       |
| `instructions-sync`        | [Sync AGENTS.md with Current State](commands/instructions-sync.md)            |
| `skill-assimilate`         | [Augment Skill with Another Skill](commands/skill-assimilate.md)              |
| `skill-create`             | [Create Guideline Skill from Document](commands/skill-create.md)              |
| `skill-decompose`          | [Decompose a Skill into Composable Skills](commands/skill-decompose.md)       |
| `skill-evaluate`           | [Seed a skill's output-eval file](commands/skill-evaluate.md)                 |
| `skill-extract`            | [Extract Skill from Codebase](commands/skill-extract.md)                      |
| `skill-optimize`           | [Trim a skill to its knowledge delta and verify](commands/skill-optimize.md)  |
| `skill-simplify`           | [Condense verbose skill files](commands/skill-simplify.md)                    |
| `slashcommand-assimilate`  | [Augment Slash Command](commands/slashcommand-assimilate.md)                  |
| `slashcommand-create`      | [Create Slash Command](commands/slashcommand-create.md)                       |
| `slashcommand-distill`     | [Distill a Command](commands/slashcommand-distill.md)                         |
| `slashcommand-simplify`    | [Simplify Slash Command Documentation](commands/slashcommand-simplify.md)     |

## Validation

Run `npx moon run plugin-agentic:ci-check` from the repository root.
