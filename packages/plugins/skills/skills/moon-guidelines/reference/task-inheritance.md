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
```

**Techniques:**

## inheritedBy Matching

- `toolchain:` - Match project's language toolchain (go, typescript, rust)
- `layer:` - Match project's layer (library, application, configuration)
- `tag:` - Match any tag in project's tags array
- Multiple criteria: All must match (AND logic)

## extends Syntax

- Single file: `extends: ./tag-go.yml`
- Extends chain: Child extends parent, inherits all tasks
- Override in extending file: Redefine task to customize

## Merging

Configs merge sequentially: global → extends → local.

- **fileGroups** combine instead of replace
- **command arrays** merge (use `script:` to fully replace inherited commands)
- **args, deps, env, inputs, outputs, toolchains** merge via configurable strategies

Merge strategy options: `mergeArgs`, `mergeDeps`, `mergeEnv`, `mergeInputs`, `mergeOutputs`, `mergeToolchains`, plus `merge` as blanket default. Strategies: `append` (default — local after inherited), `prepend` (local before inherited), `replace` (local replaces inherited), `preserve` (inherited wins, ignore local).

```yaml
# Project override with command: merges → npm run lint eslint src
command: [eslint, src]

# Project override with script: replaces completely
script: npx eslint src

# Project override replacing only deps, inheriting command/options
tasks:
  npm-publish:
    deps:
    - go-build
    options:
      mergeDeps: replace
```

## Optional Dependencies

Use `optional: true` when a tag-level task depends on a task that not all inheriting projects define:

```yaml
tasks:
  npm-publish:
    deps:
    - target: ~:build
      optional: true
    - ^:npm-publish
```

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
