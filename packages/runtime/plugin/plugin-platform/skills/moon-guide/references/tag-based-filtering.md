# tag-based-filtering: Filter Tasks by Project Tags

Target every project carrying a tag with `#tag:task`. Quote it in shell (`#` is a comment char). For multiple tags or combined criteria, use `--query` (see query-language.md).

```bash
moon run '#frontend:build'                       # single tag
moon run :build --query "tags~frontend|backend"  # regex over tags
moon run :lint --query "tags!=internal"          # exclude a tag
moon run :build --query "tags~shared && projectLayer=library"
```

## Tag-based task inheritance (Moon 2.0)

Tasks defined with `inheritedBy.tag` attach to any project whose `tags` array contains that tag. A project may carry multiple tags to compose task sets (e.g. a Go+TypeScript package pulling `tag-go.yml`, `tag-typescript.yml`, `tag-npm.yml`).

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

- Project-level tasks override tag-inherited tasks; use `script:` to fully replace an inherited command (`command:` arrays merge).
- Task `deps` merge from inherited + project definitions.
