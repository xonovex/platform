# Writing Plugin

Revise prose and write technical documents, articles, news, and travel guides.

<!-- xonovex:installation:start -->

## Install

Add the Xonovex marketplace once, then install `xonovex-writing` in the selected harness. The bundle version is `5.3.0`.

### Claude Code

Install this bundle in Claude Code. The harness discovers its bundled skills and commands.

```bash
claude plugin marketplace add xonovex/platform
claude plugin install xonovex-writing@xonovex-marketplace
```

### Codex

Install this bundle in Codex. Codex loads the skills declared in its plugin manifest; Claude Code slash commands are separate components.

```bash
codex plugin marketplace add xonovex/platform
codex plugin add xonovex-writing@xonovex-marketplace
```

<!-- xonovex:installation:end -->

## Skills

Choose the skill that owns the task. The harness loads its instructions when the request matches its routing description.

| Skill | Use when |
| --- | --- |
| [writing-editorials](skills/writing-editorials/SKILL.md) | Revising articles, posts, reviews, newsletters, marketing copy, or other publication prose for voice, rhythm, register, readability, or a natural human style. |
| [writing-news](skills/writing-news/SKILL.md) | Researching and authoring current news stories or news posts, especially multilingual Markdown with structured frontmatter. |
| [writing-prose](skills/writing-prose/SKILL.md) | Revising general prose for clarity, concision, factual discipline, or consistent terminology across documents and messages. |
| [writing-technical-docs](skills/writing-technical-docs/SKILL.md) | Structuring or rewriting standalone technical explanations, reports, status updates, documentation, recommendations, or operational handoffs so readers see the outcome, decision, action, or status first. |
| [writing-travel-guides](skills/writing-travel-guides/SKILL.md) | Researching and authoring travel, destination, attraction, or venue guides with current practical information, itineraries, or multilingual publication files. |

## Commands

Use the command name exposed by the harness. Command titles use the `xonovex-writing` namespace.

| Command | Purpose |
| --- | --- |
| `content-humanize` | [Write or revise publication prose for a natural voice while preserving its facts.](commands/content-humanize.md) |
| `content-news-add` | [Research current news on the topic and produce the required bilingual content.](commands/content-news-add.md) |
| `content-travelguide-add` | [Create a verified travel guide for the topic or location in the required languages.](commands/content-travelguide-add.md) |

## Validation

Run `npx moon run plugin-writing:ci-check --force` from the repository root to check the plugin manifests, formatting, and owned skills and commands.
