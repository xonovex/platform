# query-language: Advanced Project and Task Filtering

Use Moon Query Language (MQL) to select projects/tasks by language, layer, tags, and more. Prefer `#tag:task` for a single tag; use `--query` for multi-criterion conditions.

```bash
moon query projects "language=javascript && projectLayer=library"
moon run :build --query "language=typescript"
moon run :test --query "(language=javascript || language=typescript) && projectLayer=application"
moon run :lint --query "tags~shared && projectLayer=library"
```

## Operators

`=` equals · `!=` not equals · `~` regex match · `!~` not regex · `&&` AND · `||` OR · `()` grouping.

## Fields (Moon 2.0)

| Field           | Notes                               | Example                          |
| --------------- | ----------------------------------- | -------------------------------- |
| `language`      | project language                    | `language=typescript`            |
| `projectId`     | project id (was `projectName`)      | `projectId=core`                 |
| `projectLayer`  | layer (was `projectType`)           | `projectLayer=library`           |
| `projectAlias`  | package name alias                  | `projectAlias~@scope/*`          |
| `projectSource` | source path                         | `projectSource~packages/agent/*` |
| `tags`          | project tags                        | `tags~frontend`                  |
| `taskToolchain` | task toolchain (was `taskPlatform`) | `taskToolchain=node`             |

Prefer exact `=` over regex `~` when possible.
