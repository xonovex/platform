# Writing Plugin

Revise prose and write technical documents, articles, news, and travel guides.

<!-- xonovex:installation:start -->

## Install

Add the Xonovex marketplace once, then install `xonovex-writing` in the selected harness. The bundle version is `6.0.0`.

### Claude Code

Install this bundle in Claude Code. The harness discovers its bundled skills.

```bash
claude plugin marketplace add xonovex/platform
claude plugin install xonovex-writing@xonovex-marketplace
```

### Codex

Install this bundle in Codex. Codex loads the skills declared in its plugin manifest.

```bash
codex plugin marketplace add xonovex/platform
codex plugin add xonovex-writing@xonovex-marketplace
```

<!-- xonovex:installation:end -->

## Skills

Choose the skill that owns the task. The harness loads its instructions when the request matches its routing description.

| Skill | Use when |
| --- | --- |
| [writing-editorials](skills/writing-editorials/SKILL.md) | Revising an article, post, review, newsletter, or marketing copy so it reads naturally for its audience. |
| [writing-news](skills/writing-news/SKILL.md) | Researching and writing a current news story or news post, often as multilingual Markdown with frontmatter. |
| [writing-prose](skills/writing-prose/SKILL.md) | Making existing text clearer, shorter, or more precise, such as a README, document, message, or note. |
| [writing-technical-docs](skills/writing-technical-docs/SKILL.md) | Ordering a technical explanation, report, status update, recommendation, or handoff so the outcome, decision, action, or status comes first. |
| [writing-travel-guides](skills/writing-travel-guides/SKILL.md) | Researching and writing a travel, destination, attraction, or venue guide with current practical facts. |

## Validation

Run `npx moon run plugin-writing:ci-check --force` from the repository root to check the plugin manifests, formatting, and owned skills.
