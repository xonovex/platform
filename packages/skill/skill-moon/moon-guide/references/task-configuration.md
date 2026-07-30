# task-configuration: Define Inherited Tasks

Define tasks in `.moon/tasks/*.yml` with `inputs`/`outputs`/`deps` for automatic inheritance by matching projects.

```yaml
# .moon/tasks/node.yml - inherited by all Node.js projects
tasks:
  build:
    script: npx tsc
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

## command vs script (Moon 2.0)

- `command:`: simple executable + args only, no shell features.
- `script:`: shell commands with pipes (`|`), redirects (`>`), chaining (`&&`), variable expansion. Use `script:` whenever shell features appear.

```yaml
command: [go, build, ./...]
script: mkdir -p bin && go build -o bin/app ./cmd/...
```

## Core settings

- `inputs:`: files/globs that affect caching and invalidation.
- `outputs:`: directories/files produced, cached and restored.
- `deps:`: `~:task` (same project), `^:task` (upstream), or `project:task`; use `{target, optional: true}` for deps that may not exist.
- `options:`: `runInCI`, `persistent`, `shell`, `cache`, and merge strategies (`merge`, `mergeArgs`, `mergeDeps`, `mergeEnv`, `mergeInputs`, `mergeOutputs`, `mergeToolchains`: each `append`/`prepend`/`replace`/`preserve`).
- `env:`: task environment variables.
- `toolchains:`: execution-environment constraint (replaces v1 `platform`).

## Shell & env (Moon 2.0)

- Tasks run in a shell by default (`options.shell: true`): `bash` on Unix, `pwsh` on Windows. Set `options.shell: false` to disable.
- `$VAR` / `${VAR}` → empty string if unset; `${VAR:-default}` → default if unset; `${VAR?}` → falls back to the literal syntax if unset.
