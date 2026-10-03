# task-inheritance: Task Inheritance Hierarchy

Compose tasks with `inheritedBy` (attach to matching projects) and `extends` (chain config files), merged with Moon 2.0 deep merging.

```yaml
# .moon/tasks/tag-go.yml - tag-based mixin
tasks:
  go-build:
    command: [go, build, ./...]
    inputs: ["**/*.go", go.mod]
inheritedBy:
  tag: go

# .moon/tasks/go-library.yml - extends the mixin, adds layer-specific tasks
extends: ./tag-go.yml
tasks:
  ci-check:
    script: echo 'ci-check complete'
    deps: [go-build, go-test, go-lint]
inheritedBy:
  toolchain: go
  layer: library
```

## inheritedBy matching

- `toolchain:`: project's language toolchain (go, typescript, rust).
- `layer:`: project's layer (library, application, configuration).
- `tag:`: any tag in the project's `tags` array.
- Multiple criteria = AND (all must match).

## extends & merging

Configs merge sequentially: global → extends → local.

- `extends: ./tag-go.yml` inherits all tasks; redefine a task in the extending file to override.
- **fileGroups** and **command arrays** merge (use `script:` to fully replace an inherited command).
- `args`, `deps`, `env`, `inputs`, `outputs`, `toolchains` merge via strategy: `append` (default, local after inherited), `prepend` (local before), `replace` (local replaces), `preserve` (inherited wins). Set per-key (`mergeDeps`, `mergeArgs`, ...) or blanket via `merge`.

```yaml
# override deps only, inheriting command/options
tasks:
  npm-publish:
    deps: [go-build]
    options:
      mergeDeps: replace
```

## Optional dependencies

Use `optional: true` when a tag-level task depends on a task that not every inheriting project defines:

```yaml
tasks:
  npm-publish:
    deps:
      - target: ~:build
        optional: true
      - ^:npm-publish
```
