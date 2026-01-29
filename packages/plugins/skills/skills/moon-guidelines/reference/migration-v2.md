# migration-v2: Migrating from Moon v1 to v2

**Guideline:** Run `moon migrate v2` to automate most changes, then manually address breaking changes.

**Rationale:** Moon 2.0 introduces breaking changes to task configuration, environment variables, and file structure that require migration.

**Reference:** [Moon 2.0 Migration Guide](https://moonrepo.dev/docs/migrate/2.0)

## Automated Migration

```bash
moon migrate v2
```

This handles most configuration renames automatically.

## Breaking Changes Checklist

### Task Configuration

| v1                                     | v2                    |
| -------------------------------------- | --------------------- |
| Complex `command:` with shell features | Use `script:` instead |
| `platform: node`                       | `toolchains: [node]`  |
| `tasks.*.local: true`                  | `preset: 'server'`    |

```yaml
# v1 - BROKEN in v2
command: 'echo "foo" && echo "bar"'

# v2 - Use script for shell features
script: 'echo "foo" && echo "bar"'
```

### Environment Variables

| v1 Syntax         | v1 Behavior          | v2 Behavior          |
| ----------------- | -------------------- | -------------------- |
| `$VAR`            | Keep syntax if empty | **Empty string**     |
| `${VAR}`          | Keep syntax if empty | **Empty string**     |
| `${VAR?}`         | Empty string         | Keep syntax if empty |
| `${VAR:-default}` | Not supported        | **Use default**      |

### File Renames

| v1                    | v2                                     |
| --------------------- | -------------------------------------- |
| `.moon/toolchain.yml` | `.moon/toolchains.yml` (plural)        |
| `.moon/tasks.yml`     | `.moon/tasks/all.yml` (no inheritedBy) |

### Setting Renames

| v1              | v2               |
| --------------- | ---------------- |
| `type: library` | `layer: library` |
| `project.name`  | `project.title`  |
| `runner:`       | `pipeline:`      |
| `vcs.manager`   | `vcs.client`     |
| `$projectName`  | `$projectTitle`  |
| `$projectType`  | `$projectLayer`  |
| `$taskPlatform` | `$taskToolchain` |

### Query Language (MQL)

| v1                    | v2                     |
| --------------------- | ---------------------- |
| `projectName=foo`     | `projectId=foo`        |
| `projectType=library` | `projectLayer=library` |
| `taskPlatform=node`   | `taskToolchain=node`   |

### CLI Changes

| v1                      | v2                                            |
| ----------------------- | --------------------------------------------- |
| `--logLevel`            | `--log-level` (kebab-case)                    |
| `--update-cache`        | `--force`                                     |
| `--platform`            | `--toolchain`                                 |
| `moon run --dependents` | `moon run --dependents=deep` (value required) |

### Removed Features

- `moon node` command
- `moon migrate from-package-json` command
- `moon query hash` / `moon query hash-diff` commands
- `toolchain.*.disabled` setting (use `null` instead)
- `project.metadata` (move fields to `project` root)

## New Features in v2

- **Deep merging** - fileGroups combine instead of replace
- **Shell by default** - Tasks run in shell (`bash`/`pwsh`)
- **Default values** - `${VAR:-default}` syntax supported
- **Extensions file** - `.moon/extensions.yml` for built-in extensions
- **.env deferred loading** - Loaded just before execution

## Post-Migration Verification

```bash
# Verify all projects detected
moon query projects

# Check specific task configuration
moon task project:taskname

# Run CI prepare to verify task graph
moon run :ci-prepare --dry-run
```
