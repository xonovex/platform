# task-configuration: Define Inherited Tasks

**Guideline:** Define tasks in `.moon/tasks/*.yml` files with inputs, outputs, and dependencies for automatic inheritance by projects.

**Rationale:** Centralized task configuration prevents duplication and ensures consistent task execution across projects of the same type or language.

**Example:**

```yaml
# .moon/tasks/node.yml - Inherited by all Node.js projects
tasks:
  build:
    command: 'vite build'
    inputs: [src/**, vite.config.ts]
    outputs: [dist]

  typecheck:
    command: 'tsc --noEmit'
    inputs: [src/**, tsconfig.json]
    deps: ['~:build']

  test:
    command: 'vitest run'
    inputs: [src/**, vitest.config.ts]
    options:
      runInCI: true
```

**Techniques:**
- command: Shell command to execute
- inputs: Files/globs affecting caching and task invalidation
- outputs: Directories/files produced, cached and restored
- deps: Task dependencies using ~:task syntax
- options: runInCI, persistent, optional settings
- env: Environment variables for task execution
- platform: Constraint (node, system) for execution environment
- Inheritance: Tasks inherit to projects based on language/type matching
- Override: Project moon.yml can override or extend inherited tasks
