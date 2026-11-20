# Integrations Plugin

Use GitHub and GitLab for issues, pull requests, delivery, and continuous integration.

This plugin requires [xonovex-core](../plugin-core/README.md).

<!-- xonovex:installation:start -->

## Install

Add the Xonovex marketplace once, then install `xonovex-integrations` in the selected harness. The bundle version is `6.0.0`.

### Claude Code

Install this bundle in Claude Code. The harness discovers its bundled skills.

```bash
claude plugin marketplace add xonovex/platform
claude plugin install xonovex-integrations@xonovex-marketplace
```

### Codex

Install this bundle in Codex. Codex loads the skills declared in its plugin manifest.

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
| [using-github](skills/using-github/SKILL.md) | Doing the GitHub host mechanics of a task: issues, Projects boards, fetching a pull request, reading its review threads, posting review comments, or repository CI and enforcement. |
| [using-gitlab](skills/using-gitlab/SKILL.md) | Doing the GitLab host mechanics of a task: issues and work items, issue boards, merge-request delivery and review notes, or CI and policy enforcement. |

## Validation

Run `npx moon run plugin-integrations:ci-check --force` from the repository root to check the plugin manifests, formatting, and owned skills.
