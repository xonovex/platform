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
codex plugin add xonovex-utility@xonovex-marketplace
```

### Dependencies

Each command delegates its procedure to a guideline skill, declared in `plugin.json`
`dependencies`. On Claude Code, installing this plugin auto-installs those skills; if a
depended-on skill is missing the command is disabled with `dependency-unsatisfied`. On
Codex, `dependencies` is not auto-installed — install the delegated skill plugins
alongside this one.

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

| Command                  | Description                                          |
| ------------------------ | ---------------------------------------------------- |
| `skill-guide-create`     | Create a guideline skill from a document or URL      |
| `skill-guide-extract`    | Extract patterns from codebase into a skill          |
| `skill-guide-simplify`   | Make skills project-independent and condense         |
| `skill-guide-decompose`  | Split a multi-concern skill into single-owner skills |
| `skill-guide-assimilate` | Augment a skill with elements from another skill     |

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
