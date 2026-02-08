# query-language: Advanced Project and Task Filtering

**Guideline:** Use Moon Query Language (MQL) for advanced filtering of projects and tasks.

**Rationale:** MQL allows precise selection of projects using language, layer, tags, and other criteria for targeted task execution and querying.

**Example:**

```bash
# Query projects by language and layer (v2 syntax)
moon query projects "language=javascript && projectLayer=library"

# Run build for typescript projects
moon run :build --query "language=typescript"

# Complex filtering
moon run :test --query "(language=javascript || language=typescript) && projectLayer=application"
```

**Techniques:**

## Operators

- `=` - Equals
- `!=` - Not equals
- `~` - Regex match
- `!~` - Not regex match
- `&&` - AND
- `||` - OR
- `()` - Grouping

## Project Fields (Moon 2.0)

| Field           | Description                    | Example                          |
| --------------- | ------------------------------ | -------------------------------- |
| `language`      | Project language               | `language=typescript`            |
| `projectId`     | Project ID (was `projectName`) | `projectId=core`                 |
| `projectLayer`  | Layer type (was `projectType`) | `projectLayer=library`           |
| `projectAlias`  | Package name alias             | `projectAlias~@xonovex/*`        |
| `projectSource` | Source path                    | `projectSource~packages/agent/*` |
| `tags`          | Project tags                   | `tags~frontend`                  |

## Task Fields (Moon 2.0)

| Field           | Description                         | Example              |
| --------------- | ----------------------------------- | -------------------- |
| `taskToolchain` | Task toolchain (was `taskPlatform`) | `taskToolchain=node` |

## Usage Patterns

```bash
# Simple tag filtering
moon run '#frontend:build'

# Query with regex for multiple tags
moon run :build --query "tags~frontend|backend"

# Filter by layer
moon run :test --query "projectLayer=application"

# Complex multi-criterion
moon run :lint --query "(language=typescript && projectLayer=library) || tags~shared"
```

## Performance Tips

- Prefer exact matches (`=`) over regex (`~`) when possible
- Use `#tag:task` syntax for simple single-tag filtering
- Use `--query` for complex multi-criterion conditions
- Parentheses group conditions for boolean logic
