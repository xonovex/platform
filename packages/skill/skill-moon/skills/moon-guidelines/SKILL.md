---
name: moon-guidelines
description: Trigger on `.moon/`, `moon.yml` files in project config. Use when working with moonrepo for monorepo task management. Apply for task configuration, project tagging, task inheritance. Keywords: moon, monorepo, .moon/tasks, tags, task inheritance, project queries, task caching, language toolchains.
---

# Moon Build System Guidelines

## Requirements

- Moon ≥ 2.0, Node.js for JavaScript/TypeScript projects.
- Migration from v1: Run `moon migrate v2` to automate configuration updates.

## Essentials

- **Task configuration** - Define in `.moon/tasks/*.yml` for auto-inheritance, see [reference/task-configuration.md](reference/task-configuration.md), [reference/task-inheritance.md](reference/task-inheritance.md)
- **Tags** - Use for categorization, filtering, boundaries, inheritance, see [reference/tag-based-filtering.md](reference/tag-based-filtering.md)
- **Execution** - `project:task`, `#tag:task`, `:task` (all), `--query`, `--affected`
- **Querying** - Query projects/tasks by tags, language, or query language, see [reference/query-language.md](reference/query-language.md)
- **Configuration** - Projects use `moon.yml`, workspace uses `.moon/workspace.yml`
- **Toolchains** - Moon manages language toolchains, configured in `.moon/toolchains.yml` (plural in v2)
- **Caching** - Built-in task caching with output definitions, see [reference/task-configuration.md](reference/task-configuration.md)

## Moon 2.0 Key Changes

- **command vs script** - Use `script:` for shell features (pipes, redirects, chaining); `command:` for simple executables only
- **Shell by default** - Tasks run in shell by default (`bash` on Unix, `pwsh` on Windows)
- **Deep merging** - Configs merge sequentially (not shallow), fileGroups combine instead of replace
- **Renamed settings** - `platform` → `toolchains`, `type` → `layer`, `toolchain.yml` → `toolchains.yml`
- **Env var syntax** - `$VAR` substitutes empty string (not syntax fallback); use `${VAR:-default}` for defaults

## Progressive disclosure

- Read [reference/task-configuration.md](reference/task-configuration.md) - When defining or modifying task configurations
- Read [reference/tag-based-filtering.md](reference/tag-based-filtering.md) - When filtering projects or tasks by tags
- Read [reference/query-language.md](reference/query-language.md) - When using advanced query syntax
- Read [reference/task-inheritance.md](reference/task-inheritance.md) - When setting up task inheritance patterns
- Read [reference/project-constraints.md](reference/project-constraints.md) - When enforcing project boundaries
- Read [reference/migration-v2.md](reference/migration-v2.md) - When migrating from moon v1 to v2
- Read [reference/docker-multistage.md](reference/docker-multistage.md) - When building Docker images with moon scaffold

## External References

- [Moon 2.0 Migration Guide](https://moonrepo.dev/docs/migrate/2.0)
