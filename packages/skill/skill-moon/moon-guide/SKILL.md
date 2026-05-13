---
name: moon-guide
description: "Use when configuring moonrepo monorepo tasks. Triggers on `.moon/` config files, `moon.yml`, and prompts about task definitions, project tags, task inheritance, language toolchains, project queries, or task caching, even when the user doesn't say 'moonrepo'. Skip Nx, Turborepo, Bazel, and ad-hoc npm-script orchestration."
---

# Moon Build System Guidelines

## Requirements

- Moon ≥ 2.0, Node.js for JavaScript/TypeScript projects.
- Migration from v1: Run `moon migrate v2` to automate configuration updates.

## Essentials

- **Task configuration** - Define in `.moon/tasks/*.yml` for auto-inheritance, see [references/task-configuration.md](references/task-configuration.md), [references/task-inheritance.md](references/task-inheritance.md)
- **Tags** - Use for categorization, filtering, boundaries, inheritance, see [references/tag-based-filtering.md](references/tag-based-filtering.md)
- **Execution** - `project:task`, `#tag:task`, `:task` (all), `--query`, `--affected`
- **Querying** - Query projects/tasks by tags, language, or query language, see [references/query-language.md](references/query-language.md)
- **Configuration** - Projects use `moon.yml`, workspace uses `.moon/workspace.yml`
- **Toolchains** - Moon manages language toolchains, configured in `.moon/toolchains.yml` (plural in v2)
- **Caching** - Built-in task caching with output definitions, see [references/task-configuration.md](references/task-configuration.md)

## Moon 2.0 Key Changes

- **command vs script** - Use `script:` for shell features (pipes, redirects, chaining); `command:` for simple executables only
- **Shell by default** - Tasks run in shell by default (`bash` on Unix, `pwsh` on Windows)
- **Deep merging** - Configs merge sequentially (not shallow), fileGroups combine instead of replace
- **Renamed settings** - `platform` → `toolchains`, `type` → `layer`, `toolchain.yml` → `toolchains.yml`
- **Env var syntax** - `$VAR` substitutes empty string (not syntax fallback); use `${VAR:-default}` for defaults

## Gotchas

- Task inheritance flows from `.moon/tasks/*.yml` (by tag/language) → project `moon.yml` — overriding requires the same task key in the project file
- Implicit task dependencies via `deps:` are project-scoped — cross-project deps need `<project>:<task>` syntax
- `moon ci` skips tasks marked `local: true` — never gate CI-only checks behind a local-only task
- Project tags drive task inheritance; misspelling a tag silently disables the inherited tasks for that project

## Progressive disclosure

- Read [references/task-configuration.md](references/task-configuration.md) - Load when defining or modifying task configurations
- Read [references/tag-based-filtering.md](references/tag-based-filtering.md) - Load when filtering projects or tasks by tags
- Read [references/query-language.md](references/query-language.md) - Load when using advanced query syntax
- Read [references/task-inheritance.md](references/task-inheritance.md) - Load when setting up task inheritance patterns
- Read [references/project-constraints.md](references/project-constraints.md) - Load when enforcing project boundaries
- Read [references/migration-v2.md](references/migration-v2.md) - Load when migrating from moon v1 to v2
- Read [references/docker-multistage.md](references/docker-multistage.md) - Load when building Docker images with moon scaffold

## External References

- [Moon 2.0 Migration Guide](https://moonrepo.dev/docs/migrate/2.0)
