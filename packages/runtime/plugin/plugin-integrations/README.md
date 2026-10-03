# Integrations Plugin

Use GitHub and GitLab for issues, pull requests, delivery, and continuous integration.

This plugin requires [xonovex-core](../plugin-core/README.md).

<!-- xonovex:installation:start -->

## Install

Add the Xonovex marketplace once, then install `xonovex-integrations` in the selected harness. The bundle version is `5.3.0`.

### Claude Code

Install this bundle in Claude Code. The harness discovers its bundled skills and commands.

```bash
claude plugin marketplace add xonovex/platform
claude plugin install xonovex-integrations@xonovex-marketplace
```

### Codex

Install this bundle in Codex. Codex loads the skills declared in its plugin manifest; Claude Code slash commands are separate components.

```bash
codex plugin marketplace add xonovex/platform
codex plugin add xonovex-integrations@xonovex-marketplace
```

Plugin dependencies: `xonovex-core`.

<!-- xonovex:installation:end -->

## Skills

Choose the skill that owns the task. The harness loads its instructions when the request matches its routing description.

| Skill | Use when |
| --- | --- |
| [github-guide](skills/github-guide/SKILL.md) | Managing GitHub tickets, Projects kanban, pull-request delivery/review, durable issue or PR context comments, or native CI/repository enforcement on github.com or GitHub Enterprise Server. |
| [gitlab-guide](skills/gitlab-guide/SKILL.md) | Managing GitLab tickets/work items, issue-board kanban, merge-request delivery/review, durable issue or MR context notes, or native CI/policy enforcement on GitLab.com, Self-Managed, or Dedicated. |

## Validation

Run `npx moon run plugin-integrations:ci-check --force` from the repository root to check the plugin manifests, formatting, and owned skills.
