# migration-v2: Migrating from Moon v1 to v2

Run `moon migrate v2` to automate most config renames, then manually address the breaking changes below. See the [Moon 2.0 Migration Guide](https://moonrepo.dev/docs/migrate/2.0).

## Task configuration

| v1                                     | v2                    |
| -------------------------------------- | --------------------- |
| complex `command:` with shell features | use `script:` instead |
| `platform: node`                       | `toolchains: [node]`  |
| `tasks.*.local: true`                  | `preset: 'server'`    |

```yaml
# v1 (BROKEN in v2)
command: 'echo "foo" && echo "bar"'
# v2
script: 'echo "foo" && echo "bar"'
```

## Environment variables

| Syntax            | v1 behavior          | v2 behavior          |
| ----------------- | -------------------- | -------------------- |
| `$VAR`            | keep syntax if empty | **empty string**     |
| `${VAR}`          | keep syntax if empty | **empty string**     |
| `${VAR?}`         | empty string         | keep syntax if empty |
| `${VAR:-default}` | not supported        | **use default**      |

## File renames

| v1                    | v2                                     |
| --------------------- | -------------------------------------- |
| `.moon/toolchain.yml` | `.moon/toolchains.yml` (plural)        |
| `.moon/tasks.yml`     | `.moon/tasks/all.yml` (no inheritedBy) |

## Setting renames

| v1              | v2               |
| --------------- | ---------------- |
| `type: library` | `layer: library` |
| `project.name`  | `project.title`  |
| `runner:`       | `pipeline:`      |
| `vcs.manager`   | `vcs.client`     |
| `$projectName`  | `$projectTitle`  |
| `$projectType`  | `$projectLayer`  |
| `$taskPlatform` | `$taskToolchain` |

## Query language (MQL)

| v1                    | v2                     |
| --------------------- | ---------------------- |
| `projectName=foo`     | `projectId=foo`        |
| `projectType=library` | `projectLayer=library` |
| `taskPlatform=node`   | `taskToolchain=node`   |

## CLI changes

| v1                      | v2                                            |
| ----------------------- | --------------------------------------------- |
| `--logLevel`            | `--log-level` (kebab-case)                    |
| `--update-cache`        | `--force`                                     |
| `--platform`            | `--toolchain`                                 |
| `moon run --dependents` | `moon run --dependents=deep` (value required) |

## Removed features

- `moon node`, `moon migrate from-package-json`, `moon query hash` / `moon query hash-diff` commands.
- `toolchain.*.disabled` setting (use `null` instead).
- `project.metadata` (move fields to `project` root).

## Verify after migrating

```bash
moon query projects              # all projects detected
moon task project:taskname       # inspect a task's config
moon run :ci-check --dry-run     # verify task graph
```
