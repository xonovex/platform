---
name: moon-guidelines
description: >-
  Trigger on `.moon/`, `moon.yml` files in project config. Use when working with moonrepo for monorepo task management. Apply for task configuration, project tagging, task inheritance. Keywords: moon, monorepo, .moon/tasks, tags, task inheritance, project queries, task caching, language toolchains.
---

# Moon Build System Guidelines

## Requirements

- Moon ≥ 1.0, Node.js for JavaScript/TypeScript projects.

## Essentials

- **Task configuration** - Define in `.moon/tasks/*.yml` for auto-inheritance, see [reference/task-configuration.md](reference/task-configuration.md), [reference/task-inheritance.md](reference/task-inheritance.md)
- **Tags** - Use for categorization, filtering, boundaries, inheritance, see [reference/tag-based-filtering.md](reference/tag-based-filtering.md)
- **Execution** - `project:task`, `#tag:task`, `:task` (all), `--query`, `--affected`
- **Querying** - Query projects/tasks by tags, language, or query language, see [reference/query-language.md](reference/query-language.md)
- **Configuration** - Projects use `moon.yml`, workspace uses `.moon/workspace.yml`
- **Toolchain** - Moon manages language toolchains, configured in workspace
- **Caching** - Built-in task caching with output definitions, see [reference/task-configuration.md](reference/task-configuration.md)

## Progressive disclosure

- Read [reference/task-configuration.md](reference/task-configuration.md) - When defining or modifying task configurations
- Read [reference/tag-based-filtering.md](reference/tag-based-filtering.md) - When filtering projects or tasks by tags
- Read [reference/query-language.md](reference/query-language.md) - When using advanced query syntax
- Read [reference/task-inheritance.md](reference/task-inheritance.md) - When setting up task inheritance patterns
- Read [reference/project-constraints.md](reference/project-constraints.md) - When enforcing project boundaries
- Read [reference/docker-multistage.md](reference/docker-multistage.md) - When building Docker images with moon scaffold
