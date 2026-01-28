# task-inheritance: Task Inheritance Hierarchy

**Guideline:** Define tasks at appropriate levels using `inheritedBy` and `extends` with deep merging in Moon 2.0.

**Rationale:** Hierarchical task inheritance allows global defaults with project-specific customizations, reducing configuration duplication while maintaining flexibility.

**Example:**

```yaml
# .moon/tasks/tag-go.yml - Tag-based mixin
tasks:
  go-build:
    command: [go, build, ./...]
    inputs: ['**/*.go', go.mod]
inheritedBy:
  tag: go

# .moon/tasks/go-library.yml - Extends mixin, adds layer-specific tasks
extends: ./tag-go.yml
tasks:
  ci-prepare:
    script: echo 'ci-prepare complete'
    deps: [go-build, go-test, go-lint]
inheritedBy:
  toolchain: go
  layer: library

# moon.yml - Project override using script to prevent merging
tasks:
  ci-prepare:
    script: echo 'custom ci-prepare'
    deps: [go-build, build, lint]
```

**Techniques:**

## inheritedBy Matching
- `toolchain:` - Match project's language toolchain (go, typescript, rust)
- `layer:` - Match project's layer (library, application, configuration)
- `tag:` - Match any tag in project's tags array
- Multiple criteria: All must match (AND logic)

```yaml
inheritedBy:
  toolchain: typescript
  layer: library
```

## Deep Merging (Moon 2.0)
- Configs merge sequentially: global → extends → local
- **fileGroups combine** instead of replace (both sources merge)
- **deps arrays merge** from inherited + project tasks
- **command arrays merge** (use `script:` to fully override)

## Preventing Merge Issues
- Use `script:` instead of `command:` to fully replace inherited commands
- Project-level `script:` completely overrides template task

```yaml
# Template defines:
command: [npm, run, lint]

# Project override with command: merges → npm run lint eslint src
command: [eslint, src]

# Project override with script: replaces completely
script: npx eslint src
```

## extends Syntax
- Single file: `extends: ./tag-go.yml`
- Extends chain: Child extends parent, inherits all tasks
- Override in extending file: Redefine task to customize

## Composition Patterns

**Tag-based mixins for mixed packages:**
```yaml
# Project with Go + TypeScript
language: go
layer: library
tags: [go, typescript, npm]
tasks:
  ci-prepare:
    script: echo 'done'
    deps: [go-build, go-test, build, lint, typecheck]
```

**Layer templates extending tag mixins:**
```yaml
# .moon/tasks/typescript-library.yml
extends: ./tag-typescript.yml
tasks:
  ci-prepare:
    deps: [build, test, lint, typecheck]
inheritedBy:
  toolchain: typescript
  layer: library
```
