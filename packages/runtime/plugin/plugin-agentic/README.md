# Agentic Plugin

Configure coding-agent harnesses and author their skills, commands, and project instructions.

## Install

Add the [Xonovex marketplace](../../../../README.md#agent-plugins) once, then install this plugin in the selected harness.

```bash
# Claude Code
claude plugin install xonovex-agentic@xonovex-marketplace

# Codex
codex plugin add xonovex-agentic@xonovex-marketplace
```

## Skills

Choose the skill that owns the task. The harness loads its instructions when the request matches its routing description.

| Skill | Use when |
| --- | --- |
| [claude-code-guide](skills/claude-code-guide/SKILL.md) | Configuring Claude Code hooks, plugins, skills, settings, or managed configuration. |
| [codex-guide](skills/codex-guide/SKILL.md) | Configuring Codex hooks, plugins, skills, config layers, or managed requirements. |
| [command-guide](skills/command-guide/SKILL.md) | Authoring, reviewing, merging, simplifying, or distilling reusable user-invocable prompt files (a.k.a. slash commands: files an agent harness exposes as `/command` invocations). |
| [copilot-guide](skills/copilot-guide/SKILL.md) | Configuring GitHub Copilot CLI or cloud-agent hooks, policy hooks, plugins, or skills. |
| [instruction-guide](skills/instruction-guide/SKILL.md) | Authoring, reviewing, initializing, syncing, simplifying, consolidating, or assimilating AGENTS.md project-instruction files. |
| [kiro-guide](skills/kiro-guide/SKILL.md) | Configuring Kiro IDE or CLI v3 hooks, command actions, or agent actions. |
| [llmstxt-guide](skills/llmstxt-guide/SKILL.md) | Authoring, reviewing, or maintaining an `/llms.txt` file or per-page markdown mirrors per the llmstxt.org specification. |
| [opencode-guide](skills/opencode-guide/SKILL.md) | Configuring OpenCode JavaScript or TypeScript plugins, events, custom tools, or settings. |
| [pi-guide](skills/pi-guide/SKILL.md) | Configuring Pi extensions, packages, skills, settings, context injection, tool interception, permissions, or subagent patterns. |
| [skill-guide](skills/skill-guide/SKILL.md) | Authoring, reviewing, extracting, merging, simplifying, decomposing, or validating Agent Skills (SKILL.md plus references / scripts / assets), or when auditing, splitting, de-duplicating, or tiering a set of skills. |

## Commands

Use the command name exposed by the harness. Command titles use the `xonovex-agentic` namespace.

| Command | Purpose |
| --- | --- |
| `instructions-assimilate` | [Add useful guidance from another project while preserving the target instructions and their structure.](commands/instructions-assimilate.md) |
| `instructions-consolidate` | [Remove redundant instruction files and keep each rule in the directory that owns it.](commands/instructions-consolidate.md) |
| `instructions-init` | [Create AGENTS.md from the directory structure, tools, and project conventions.](commands/instructions-init.md) |
| `instructions-simplify` | [Shorten AGENTS.md while preserving its rules, commands, and project context.](commands/instructions-simplify.md) |
| `instructions-sync` | [Update AGENTS.md to match the current directory structure and project state.](commands/instructions-sync.md) |
| `skill-assimilate` | [Add useful guidance from another skill while preserving the target skill and its structure.](commands/skill-assimilate.md) |
| `skill-create` | [Create a reusable guideline skill from the supplied document or URL.](commands/skill-create.md) |
| `skill-decompose` | [Split a skill into focused skills with one owner per concept and explicit cross-references.](commands/skill-decompose.md) |
| `skill-evaluate` | [Create or refresh output evaluations with realistic prompts, binary assertions, and explicit tiers.](commands/skill-evaluate.md) |
| `skill-extract` | [Create or update a skill from recurring code and project-instruction patterns.](commands/skill-extract.md) |
| `skill-optimize` | [Keep the guidance the weakest target model needs and verify each removal through evaluation.](commands/skill-optimize.md) |
| `skill-simplify` | [Remove redundant guidance and move detailed examples into references while preserving the skill behavior.](commands/skill-simplify.md) |
| `slashcommand-assimilate` | [Add useful guidance from another command while preserving the target argument contract and delegation.](commands/slashcommand-assimilate.md) |
| `slashcommand-create` | [Create a thin command that defines its arguments and delegates its procedure to one owning skill.](commands/slashcommand-create.md) |
| `slashcommand-distill` | [Move a command procedure into one owning skill and keep the command as a thin interface.](commands/slashcommand-distill.md) |
| `slashcommand-simplify` | [Shorten a slash command while preserving its arguments and behavior.](commands/slashcommand-simplify.md) |

## Validation

Run `npx moon run plugin-agentic:ci-check --force` from the repository root to check the plugin manifests, formatting, and owned skills and commands.
