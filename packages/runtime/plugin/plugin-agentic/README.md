# Agentic Plugin

Configure coding-agent harnesses and author their skills, commands, and project instructions.

<!-- xonovex:installation:start -->

## Install

Add the Xonovex marketplace once, then install `xonovex-agentic` in the selected harness. The bundle version is `5.3.0`.

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
| [authoring-commands](skills/authoring-commands/SKILL.md) | Authoring, reviewing, merging, simplifying, or distilling reusable user-invocable prompt files (a.k.a. slash commands: files an agent harness exposes as `/command` invocations). |
| [authoring-instructions](skills/authoring-instructions/SKILL.md) | Authoring, reviewing, initializing, syncing, simplifying, consolidating, or assimilating AGENTS.md project-instruction files. |
| [authoring-llmstxt](skills/authoring-llmstxt/SKILL.md) | Authoring, reviewing, or maintaining an `/llms.txt` file or per-page markdown mirrors per the llmstxt.org specification. |
| [authoring-skills](skills/authoring-skills/SKILL.md) | Authoring, reviewing, extracting, merging, simplifying, decomposing, or validating Agent Skills (SKILL.md plus references / scripts / assets), or when auditing, splitting, de-duplicating, or tiering a set of skills. |
| [configuring-claude-code](skills/configuring-claude-code/SKILL.md) | Configuring Claude Code hooks, plugins, skills, settings, or managed configuration. |
| [configuring-codex](skills/configuring-codex/SKILL.md) | Configuring Codex hooks, plugins, skills, config layers, or managed requirements. |
| [configuring-copilot](skills/configuring-copilot/SKILL.md) | Configuring GitHub Copilot CLI or cloud-agent hooks, policy hooks, plugins, or skills. |
| [configuring-kiro](skills/configuring-kiro/SKILL.md) | Configuring Kiro IDE or CLI v3 hooks, command actions, or agent actions. |
| [configuring-opencode](skills/configuring-opencode/SKILL.md) | Configuring OpenCode JavaScript or TypeScript plugins, events, custom tools, or settings. |
| [configuring-pi](skills/configuring-pi/SKILL.md) | Configuring Pi extensions, packages, skills, settings, context injection, tool interception, permissions, or subagent patterns. |
| [evaluating-skills](skills/evaluating-skills/SKILL.md) | Measuring an Agent Skill with model runs: trigger evals, output evals with and without the skill, or an ablation after a trim. |
| [using-herdr](skills/using-herdr/SKILL.md) | Inspecting and controlling Herdr agents, terminal commands, spaces, tabs, and panes. |

## Validation

Run `npx moon run plugin-agentic:ci-check --force` from the repository root to check the plugin manifests, formatting, and owned skills.
