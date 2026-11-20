# Shared Config Go

Use this shared Go library to read the layered `agentic.json` configuration: find the managed, user and project files, check them, pass a project file through the digest approval, and merge the layers key by key.

## Layers

| Layer | File |
| --- | --- |
| managed | `/etc/agentic/agentic.json`, `/Library/Application Support/agentic/agentic.json` on macOS, `%ProgramData%\agentic\agentic.json` on Windows |
| user | `$XDG_CONFIG_HOME/agentic/agentic.json`, else `~/.config/agentic/agentic.json` (`%AppData%\agentic` on Windows) |
| project | The nearest `agentic.json` from the start directory up to the root of the git or jj repository; outside a repository only the start directory |
| flags | The command line of the run |

A file holds `$schema`, `version: 1` and the sections `run`, `providers`, `validate`, `audit` and `eval`. A tool checks the sections it owns against their schemas in `shared-contract-go` and only type-checks the other sections.

## Merge rules

| Rule | Result |
| --- | --- |
| `replace` | One value: managed, then flags, then project, then user. A flag that differs from a managed value is an error |
| `append` | The items of every layer, the managed items last |
| `require` | True when any layer sets it true |
| `cap` | A ceiling: the user or the managed file sets it, a project may only lower it, a flag may choose any value up to the managed one |
| `merge` | Each entry of a map comes whole from the layer that `replace` chooses |

A key declares the files that may set it. A project file that sets a key closed to it is an error.

## Trust gate

The `run`, `providers` and `eval` sections of a project file take effect only after the user approves the sha256 digest of the file. At a terminal the gate shows the file and asks `Approve this digest? [y/N]`; without one the user passes `--approve-config-digest <digest>`. An unapproved file fails closed. Approvals are kept in `$XDG_STATE_HOME/agentic/trust.json`.

## Packages

| Package | Description |
| --- | --- |
| `pkg/spec` | Layers, merge rules, keys and sections as data |
| `pkg/section/<name>` | The declaration of one section: its schema, its keys and whether it is gated |
| `pkg/layer` | Envelope check, sections and digest of one file |
| `pkg/discover` | The files of the layers and the directories of a host |
| `pkg/merge` | The merge of the layers with the origin of each value |
| `pkg/trust` | The digest approval gate and the approval ledger |
| `pkg/explain` | The merged values with their origins, for `config show --sources` |
| `pkg/resolve` | Find, read, check, gate and merge in one call |
| `pkg/shard` | The override files of one kind, such as `harnesses.d/<name>.json`, beside the `agentic.json` of each layer; a project shard passes the trust gate |
| `pkg/osfiles` | Adapter for the disk, the host directories and the terminal |
| `pkg/configtest` | In-memory test doubles |

Only `pkg/osfiles` touches the host, and only a composition root imports it or `pkg/configtest`. `internal/arch` enforces both, keeps each key in step with its schema, keeps each section package to its own section, and proves that a project file cannot set a key closed to it.
