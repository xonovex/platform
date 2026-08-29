# Utility Commands

Install these utilities to manage project instructions, capture session lessons, author content, and create skills or commands.

## Installation

Claude Code installs one utility plugin. Codex installs only the individual skills required for the intended operation.

### Claude Code

```bash
claude plugin marketplace add xonovex/platform
claude plugin install xonovex-utility@xonovex-marketplace
```

### Codex

```bash
codex plugin marketplace add xonovex/platform
codex plugin add xonovex-skill-editorial-writing@xonovex-marketplace
codex plugin add xonovex-skill-news-writing@xonovex-marketplace
codex plugin add xonovex-skill-travel-writing@xonovex-marketplace
codex plugin add xonovex-skill-instruction@xonovex-marketplace
codex plugin add xonovex-skill-reflect@xonovex-marketplace
codex plugin add xonovex-skill-skill@xonovex-marketplace
codex plugin add xonovex-skill-command@xonovex-marketplace
```

In Codex, invoke the matching `$...-guide` skill directly. The `/xonovex-utility:*` command namespace is available only in Claude Code.

### Claude Code dependencies

Installing the Claude Code plugin installs the guideline skills declared in `plugin.json` dependencies. A command is disabled with `dependency-unsatisfied` when a required skill is missing.

## Commands

Choose the command group that owns the required operation.

### Content

Use these commands to edit or create publication content.

| Command                   | Description                                                                 |
| ------------------------- | --------------------------------------------------------------------------- |
| `content-humanize`        | Remove AI writing patterns and add human voice                              |
| `content-news-add`        | Auto-curate latest news stories on a topic and generate bilingual content   |
| `content-travelguide-add` | Create a comprehensive, multi-language travel guide for a topic or location |

### Instructions

Use these commands to create or maintain project instruction files.

| Command                    | Description                                               |
| -------------------------- | --------------------------------------------------------- |
| `instructions-init`        | Create an AGENTS.md file by analyzing directory structure |
| `instructions-simplify`    | Reduce verbosity in AGENTS.md files                       |
| `instructions-sync`        | Sync AGENTS.md files with current directory structure     |
| `instructions-consolidate` | Remove redundant files and standardize format             |
| `instructions-assimilate`  | Augment instructions with elements from another project   |

### Reflect

Use these commands to convert session evidence into reusable guidance.

| Command                   | Description                                                  |
| ------------------------- | ------------------------------------------------------------ |
| `reflect-extract`         | Analyze session for development mistakes and lessons learned |
| `reflect-to-instructions` | Convert insights into AGENTS.md bullet points                |
| `reflect-to-skill`        | Convert insights into a progressive disclosure skill         |

### Skills

Use these commands to create and maintain Agent Skills.

| Command            | Description                                                          |
| ------------------ | -------------------------------------------------------------------- |
| `skill-create`     | Create a guideline skill from a document or URL                      |
| `skill-extract`    | Extract patterns from codebase into a skill                          |
| `skill-simplify`   | Make skills project-independent and condense                         |
| `skill-decompose`  | Split a multi-concern skill into single-owner skills                 |
| `skill-assimilate` | Augment a skill with elements from another skill                     |
| `skill-optimize`   | Trim a skill (or catalog) to its knowledge delta and ablation-verify |
| `skill-evaluate`   | Seed a skill's evals.json output-eval file                           |

### Slash Commands

Use these commands to create and maintain reusable slash commands.

| Command                   | Description                                                |
| ------------------------- | ---------------------------------------------------------- |
| `slashcommand-create`     | Create a thin command with one owner skill                 |
| `slashcommand-simplify`   | Reduce verbosity in slash command files                    |
| `slashcommand-assimilate` | Augment a slash command with elements from another         |
| `slashcommand-distill`    | Distill a fat command with explicit owner/supporting needs |
