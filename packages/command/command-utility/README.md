# Utility Commands

Manage project instructions, reflect on sessions, create skills, and bump package versions.

## Installation

### Claude Code

```bash
claude plugin marketplace add xonovex/platform
claude plugin install xonovex-utility@xonovex-marketplace
```

### Codex

```bash
codex plugin marketplace add xonovex/platform
codex plugin add xonovex-skill-content@xonovex-marketplace
codex plugin add xonovex-skill-instruction@xonovex-marketplace
codex plugin add xonovex-skill-reflect@xonovex-marketplace
codex plugin add xonovex-skill-skill@xonovex-marketplace
codex plugin add xonovex-skill-command@xonovex-marketplace
codex plugin add xonovex-skill-versioning@xonovex-marketplace
```

Install only the skills needed for the intended operation and invoke the matching
`$...-guide` skill directly. The `/xonovex-utility:*` command namespace is a Claude
Code surface.

### Claude Code dependencies

Each command delegates its procedure to a guideline skill, declared in `plugin.json`
`dependencies`. Installing this plugin auto-installs those skills; if a depended-on
skill is missing the command is disabled with `dependency-unsatisfied`.

## Commands

### Content

| Command                   | Description                                                                 |
| ------------------------- | --------------------------------------------------------------------------- |
| `content-humanize`        | Remove AI writing patterns and add human voice                              |
| `content-news-add`        | Auto-curate latest news stories on a topic and generate bilingual content   |
| `content-travelguide-add` | Create a comprehensive, multi-language travel guide for a topic or location |

### Instructions

| Command                    | Description                                               |
| -------------------------- | --------------------------------------------------------- |
| `instructions-init`        | Create an AGENTS.md file by analyzing directory structure |
| `instructions-simplify`    | Reduce verbosity in AGENTS.md files                       |
| `instructions-sync`        | Sync AGENTS.md files with current directory structure     |
| `instructions-consolidate` | Remove redundant files and standardize format             |
| `instructions-assimilate`  | Augment instructions with elements from another project   |

### Reflect

| Command                   | Description                                                  |
| ------------------------- | ------------------------------------------------------------ |
| `reflect-extract`         | Analyze session for development mistakes and lessons learned |
| `reflect-to-instructions` | Convert insights into AGENTS.md bullet points                |
| `reflect-to-skill`        | Convert insights into a progressive disclosure skill         |

### Skills

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

| Command                   | Description                                                |
| ------------------------- | ---------------------------------------------------------- |
| `slashcommand-create`     | Create a new slash command from a completed task           |
| `slashcommand-simplify`   | Reduce verbosity in slash command files                    |
| `slashcommand-assimilate` | Augment a slash command with elements from another         |
| `slashcommand-distill`    | Distill a fat command into a thin skill-delegating command |

### Versioning

| Command        | Description                                                                     |
| -------------- | ------------------------------------------------------------------------------- |
| `version-bump` | Bump a package version, propagate to dependents, and generate a changelog entry |
