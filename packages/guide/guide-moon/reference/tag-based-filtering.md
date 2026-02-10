# tag-based-filtering: Filter Tasks by Project Tags

**Guideline:** Use `#tag:task` syntax to run tasks for all projects matching a specific tag.

**Rationale:** Tag-based filtering enables efficient execution of tasks across related projects without listing each individually, supporting tenant isolation and feature organization.

**Example:**

```bash
# Run build for all frontend projects
moon run '#frontend:build'

# Run tests for tenant-specific projects
moon run '#tenant-example:test'
moon run '#tenant-other:lint'

# Multiple tags with query
moon run :build --query "tags~frontend|backend"
```

**Techniques:**

## Tag Syntax

- `#tag:task` - Target all projects with specific tag
- Quote in shell: `'#tag:task'` (# is comment character)
- Multiple tags: Use `--query "tags~tag1|tag2"`

## Tag-Based Task Inheritance (Moon 2.0)

Define tasks that inherit to projects with specific tags:

```yaml
# .moon/tasks/tag-npm.yml
tasks:
  npm-publish:
    command: [npm, publish, --provenance, --access, public]
    deps: [^:npm-publish]
    options:
      cache: false
      runInCI: false
inheritedBy:
  tag: npm
```

Projects with `tags: [npm]` automatically get `npm-publish` task.

## Composing with Multiple Tags

Mixed-language packages can use multiple tags:

```yaml
# moon.yml - Project configuration
language: go
layer: library
tags:
  - go # Gets Go tasks from tag-go.yml
  - typescript # Gets TypeScript tasks from tag-typescript.yml
  - npm # Gets npm-publish from tag-npm.yml
  - cli
```

## Project Tag Assignment

```yaml
# moon.yml
tags:
  - frontend
  - shared
  - tenant-example
```

## Filtering Patterns

```bash
# Single tag execution
moon run '#shared:build'

# Regex matching multiple tags
moon run :test --query "tags~frontend|backend"

# Exclude specific tag
moon run :lint --query "tags!=internal"

# Combine with layer
moon run :build --query "tags~shared && projectLayer=library"
```

## Override Behavior

- Project-level tasks override tag-inherited tasks
- Use `script:` to completely replace inherited commands
- Task deps merge from inherited + project definitions
