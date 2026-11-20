# Agentic Plugin

Configure coding-agent harnesses and author their skills, commands, and project instructions.

<!-- xonovex:installation:start -->

## Install

Add the Xonovex marketplace once, then install `xonovex-agentic` in the selected harness. The bundle version is `6.0.0`.

### Claude Code

Install this bundle in Claude Code. The harness discovers its bundled skills.

```bash
claude plugin marketplace add xonovex/platform
claude plugin install xonovex-agentic@xonovex-marketplace
```

### Codex

Install this bundle in Codex. Codex loads the skills declared in its plugin manifest.

```bash
codex plugin marketplace add xonovex/platform
codex plugin add xonovex-agentic@xonovex-marketplace
```

<!-- xonovex:installation:end -->

## Skills

Choose the skill that owns the task. The harness loads its instructions when the request matches its routing description.

| Skill | Use when |
| --- | --- |
| [authoring-commands](skills/authoring-commands/SKILL.md) | Writing, reviewing, merging, shortening, or distilling a slash command or a reusable prompt file. |
| [authoring-instructions](skills/authoring-instructions/SKILL.md) | Writing, reviewing, updating, shortening, merging, or porting the AGENTS.md project instructions of a repository. |
| [authoring-llmstxt](skills/authoring-llmstxt/SKILL.md) | Writing or updating an `llms.txt` file or Markdown copies of site pages for LLMs. |
| [authoring-skills](skills/authoring-skills/SKILL.md) | Creating, reviewing, extracting, splitting, merging, shortening, or validating an Agent Skill or a set of skills. |
| [configuring-claude-code](skills/configuring-claude-code/SKILL.md) | Setting up, testing, or debugging Claude Code hooks, plugins, skills, settings, or managed configuration. |
| [configuring-codex](skills/configuring-codex/SKILL.md) | Setting up or debugging Codex hooks, plugins, skills, config layers, or managed requirements. |
| [configuring-copilot](skills/configuring-copilot/SKILL.md) | Setting up or debugging GitHub Copilot CLI or cloud-agent hooks, policy hooks, plugins, or skills. |
| [configuring-harness-integrations](skills/configuring-harness-integrations/SKILL.md) | Adding, reviewing, comparing, or removing a guardrail, hook, plugin, or extension for an agent harness, whichever harness it is. |
| [configuring-kiro](skills/configuring-kiro/SKILL.md) | Setting up or debugging Kiro IDE or CLI v3 hooks, command actions, or agent actions. |
| [configuring-opencode](skills/configuring-opencode/SKILL.md) | Writing or debugging OpenCode JavaScript or TypeScript plugins, events, custom tools, or settings. |
| [configuring-pi](skills/configuring-pi/SKILL.md) | Writing or configuring Pi extensions, packages, skills, settings, or permissions. |
| [evaluating-skills](skills/evaluating-skills/SKILL.md) | Testing by model runs if a skill triggers or helps: trigger evals, output evals with and without it, or a trim ablation. |
| [using-herdr-tui](skills/using-herdr-tui/SKILL.md) | Using the Herdr terminal multiplexer and its `herdr` CLI to launch, inspect, or control agents, panes, tabs, workspaces, and terminal commands in a Herdr TUI session. |

## Validation

Run `npx moon run plugin-agentic:ci-check --force` from the repository root to check the plugin manifests, formatting, and owned skills.
