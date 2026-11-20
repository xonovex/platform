# Agent Audit Go

Use this CLI to audit the source catalogs of skills and documents, compare a catalog with a base ref, and convert a `SOURCES.md` into `sources.json`.

## Commands

| Command | Does |
| --- | --- |
| `agent-audit sources [path...]` | Audits every catalog below each path, the current directory by default |
| `agent-audit convert <skill-directory \| SOURCES.md>` | Writes `sources.json` beside a `SOURCES.md` without losing a word |
| `agent-audit completion <shell>` | Prints a shell completion script |

### Audit options

| Option | Effect |
| --- | --- |
| `--max-age DAYS` | Days after which a review is overdue, 180 by default |
| `--version-max-age DAYS` | The window for a source that pins a version, 90 by default |
| `--fetch` | Reads the documents of web and http sources |
| `--pull` | Refreshes the tags of each checkout before the comparison |
| `--mark-reviewed` | Sets the review date of the sources to today, limited by `--match TITLE` |
| `--base REF` | Compares each catalog with the catalog at a git ref |
| `--cover-dir DIR` | Directory whose Markdown files each need a source, `references` by default |
| `--format text\|json\|sarif`, `--json` | Output format |
| `--strict` | Fails on a warning as well |

The exit status is 0 when nothing reaches the failing level, 1 when a finding does, and 2 when the run could not complete.

## Findings

| Code | Level | Meaning |
| --- | --- | --- |
| `audit.review-overdue` | fail | The review window ran out |
| `audit.anchor-missing` | drift | A claim names an anchor that the source no longer holds |
| `audit.metadata-changed` | drift | A digest or pin no longer matches, or changed against the base ref without a review |
| `audit.git-file-changed` | drift | Watched files changed since the pin, or fed files changed without a review |
| `audit.behind-latest-tag` | warn | The pinned version is older than the latest release |
| `audit.provider-error`, `audit.provider-restricted` | fail, warn | A provider failed or refused to answer |
| `audit.reference-missing`, `audit.reference-dangling`, `audit.reference-uncovered` | fail | The mapping from sources to reference files is incomplete |
| `audit.source-removed`, `audit.review-regressed` | warn, fail | A comparison against the base ref found a dropped source or a review date that moved back |

## Layout

| Package | Concern |
| --- | --- |
| `cmd/agent-audit` | Composition root |
| `internal/cmd` | Command tree, the driving adapter |
| `internal/audit` | Composes the rules with the transports for one catalog |
| `internal/rules` | Pure checks of review age, mapping, pins, anchors and observations |
| `internal/observe` | Transport port and registry |
| `internal/transport/{filesystem,git,http}` | Transports, bound in `internal/transport/plugins` |
| `internal/compare`, `internal/compare/gitbase`, `internal/gitfs` | Comparison with a base ref |
| `internal/stamp` | Writes review dates |
| `internal/legacy` | `SOURCES.md` converter |
| `internal/render` | Text, JSON and SARIF output |
| `internal/process`, `internal/web` | Ports for programs and HTTP, with `osexec`, `nethttp` and test doubles |

## Tasks

| Task | Runs |
| --- | --- |
| `npx moon run agent-audit-go:ci-check` | Build, lint, vet, unit and architecture tests, coverage floors, format check |
| `npx moon run agent-audit-go:ci-integration` | The built binary against a real git repository and `httptest`, and the validation of every `sources.json` of the plugins |
