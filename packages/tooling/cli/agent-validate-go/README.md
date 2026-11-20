# Agent Validate Go

Use this command to validate skills, plugins, marketplaces, instructions, hooks, commands, agents and repository files. It is a Go binary that reads the repository and reports findings as text or SARIF.

## Installation

Install the command globally from npm. The package selects the binary for the platform; the supported platforms are Linux and macOS on x64 and arm64, and Windows on x64.

```bash
npm install -g @xonovex/agent-validate-go
```

## Usage

```bash
agent-validate                       # every check the validate section enables
agent-validate skill text            # the named checks only
agent-validate text --path docs      # one check over one path
agent-validate --format sarif        # SARIF 2.1.0 on standard output
agent-validate --strict              # a warning fails the run too
agent-validate list                  # the checks and what each one does
agent-validate seed-budgets          # record each skill file's word count in budgets.json
agent-validate config show --sources # the merged validate section and the file of each value
```

| Flag | Description |
| --- | --- |
| `--root` | Repository root to validate, default the working directory |
| `--project` | Project `agentic.json` relative to the root, default the nearest `agentic.json` from the root up to the repository root |
| `--format` | `text` or `sarif` |
| `--strict` | Fail on warnings as well as on failures and drift |
| `--path` | Path to check instead of the paths of the validate section, repeatable |

`seed-budgets` takes `--path` (default `packages/runtime/plugin`) and `--file` (default `budgets.json`) and rewrites the manifest with the current prose word count of every `SKILL.md` and reference file. Review the diff: a lower count tightens a budget, and a higher count is a budget bump that a reviewer should see.

Exit status 0 means the run is clean. Exit status 1 means a check reported a failure or drift, or a warning under `--strict`. Exit status 2 means a usage error or a fault of the run, such as a file that cannot be read.

## Checks

| Check | Description |
| --- | --- |
| `skill` | Agent Skills specification and catalog rules for each `SKILL.md`, its references and its evals |
| `sources` | Skill source catalogs against `sources.schema.json` and the rules across sources |
| `drift` | Word budgets per file and invariants restated across skill files |
| `links` | Relative Markdown links and their heading anchors, and bold skill handoffs |
| `routing` | Validation routing scenarios: one owner per query and cited operations that exist |
| `plugin` | Plugin manifests, skill packaging, dependencies, README catalog and handoffs |
| `marketplace` | Claude and Codex marketplaces against the plugin packages and the release version |
| `instructions` | Eval corpora of `AGENTS.md` files: schema, anchors and rule coverage |
| `command` | Slash command frontmatter, argument hints and body length |
| `agent` | Agent definition frontmatter and prompts |
| `output-style` | Output style frontmatter and instructions |
| `hook` | Hook event groups and handler configuration |
| `kit` | Docker Sandbox kit v2 grammar and the rules between its fields |
| `llms` | `llms.txt` indexes whose Markdown targets exist |
| `schema` | JSON documents against the contract schemas, by rule or by their `$schema` member |
| `repository` | `AGENTS.md` and `CLAUDE.md` pairs, README claims and release workflow safety |
| `workspace` | Moon task inheritance, coverage floors, dependency sources, Go dependencies, Go coverage floors (`GO_COVERAGE_MIN` and a `GO_PACKAGE_FLOORS` entry for every package with Go code and for no other path), hasher ignores and bin targets |
| `text` | Typographic punctuation in prose files and hard-wrapped Markdown |

## Configuration

The `validate` section of `agentic.json` selects checks, paths, options and severity. Every key is optional. `shared-config-go` merges the section of the managed, user and project files: `skip` collects the entries of every file, and each entry of `checks` comes whole from the file with the highest precedence (managed, then project, then user). The section is not gated, so a project file takes effect without a digest approval.

```json
{
  "$schema": "https://schemas.xonovex.com/agent/agentic.schema.json",
  "version": 1,
  "validate": {
    "skip": ["vendor"],
    "checks": {
      "skill": { "options": { "vendorAdapters": ["configuring-claude-code"], "descriptionLimit": 820, "bodyFailLines": 500 } },
      "text": { "level": "warn", "paths": ["docs"] },
      "kit": { "enabled": false }
    }
  }
}
```

A `level` of `warn` or `drift` lowers the severity of every finding of that check to the cap. A check that is named on the command line runs even when the configuration switches it off. The limits of a check, such as the 820-character description and the 500-line body of the `skill` check, are options with those defaults. An option that a check does not know is an error, so a misspelled option never goes unnoticed.

## Development

```bash
npx moon run agent-validate-go:ci-check          # unit tests, architecture test, coverage floors, lint
npx moon run agent-validate-go:ci-integration    # builds the binary and runs it over this repository
```
