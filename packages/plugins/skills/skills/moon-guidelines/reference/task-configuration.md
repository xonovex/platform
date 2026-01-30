# task-configuration: Define Inherited Tasks

**Guideline:** Define tasks in `.moon/tasks/*.yml` files with inputs, outputs, and dependencies for automatic inheritance by projects.

**Rationale:** Centralized task configuration prevents duplication and ensures consistent task execution across projects of the same type or language.

**Example:**

```yaml
# .moon/tasks/node.yml - Inherited by all Node.js projects
tasks:
  build:
    script: npx tsc # Use script for shell execution
    inputs: [src/**, tsconfig.json]
    outputs: [dist]

  typecheck:
    script: npx tsc --noEmit
    inputs: [src/**, tsconfig.json]
    deps: [~:build]

  test:
    script: npx vitest run
    inputs: [src/**, vitest.config.ts]
    options:
      runInCI: true
```

**Techniques:**

## command vs script (Moon 2.0)

- `command:` - Simple executables with arguments only (no shell features)
- `script:` - Shell commands with pipes (`|`), redirects (`>`), chaining (`&&`), variable expansion
- **Rule:** If command uses shell features, use `script:`

```yaml
# Simple command - use command:
command: [go, build, ./...]

# Shell features - use script:
script: mkdir -p bin && go build -o bin/app ./cmd/...
```

## Core Settings

- `inputs:` - Files/globs affecting caching and task invalidation
- `outputs:` - Directories/files produced, cached and restored
- `deps:` - Task dependencies using `~:task` (same project), `^:task` (upstream), or `project:task` syntax; use `{target, optional: true}` for deps that may not exist
- `options:` - runInCI, persistent, shell, cache, merge strategy settings (`merge`, `mergeArgs`, `mergeDeps`, `mergeEnv`, `mergeInputs`, `mergeOutputs`, `mergeToolchains` — each supports `append`/`prepend`/`replace`/`preserve`)
- `env:` - Environment variables for task execution
- `toolchains:` - Constraint for execution environment (replaces `platform` in v2)

## Shell Execution (Moon 2.0 defaults)

- Tasks run in shell by default (`options.shell: true`)
- Unix: `bash`, Windows: `pwsh`
- Disable with `options.shell: false` for non-shell execution

## Environment Variables (Moon 2.0)

- `$VAR` / `${VAR}` - Substitutes with empty string if unset
- `${VAR:-default}` - Use default value if unset
- `${VAR?}` - Fall back to literal syntax if unset

## Inheritance

- Tasks inherit to projects based on `inheritedBy` matching (toolchain, layer, tag)
- Project `moon.yml` can override or extend inherited tasks
- Use `script:` in overrides to prevent command merging issues
