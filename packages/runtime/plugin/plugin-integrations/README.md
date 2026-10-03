# Integrations Plugin

Use GitHub and GitLab for issues, pull requests, delivery, and continuous integration.

This plugin requires [xonovex-core](../plugin-core/README.md).

## Install

Add the [Xonovex marketplace](../../../../README.md#agent-plugins) once, then install this plugin in the selected harness.

```bash
# Claude Code
claude plugin install xonovex-integrations@xonovex-marketplace

# Codex
codex plugin add xonovex-integrations@xonovex-marketplace
```

## Skills

Choose the skill that owns the task. The harness loads its instructions when the request matches its routing description.

| Skill | Use when |
| --- | --- |
| [github-guide](skills/github-guide/SKILL.md) | Managing GitHub tickets, Projects kanban, pull-request delivery/review, durable issue or PR context comments, or native CI/repository enforcement on github.com or GitHub Enterprise Server. |
| [gitlab-guide](skills/gitlab-guide/SKILL.md) | Managing GitLab tickets/work items, issue-board kanban, merge-request delivery/review, durable issue or MR context notes, or native CI/policy enforcement on GitLab.com, Self-Managed, or Dedicated. |

## Validation

Run `npx moon run plugin-integrations:ci-check --force` from the repository root to check the plugin manifests, formatting, and owned skills.
