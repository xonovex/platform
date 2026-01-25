# task-inheritance: Task Inheritance Hierarchy

**Guideline:** Define tasks at appropriate levels (global, type, language, tag, project) with later definitions overriding earlier ones.

**Rationale:** Hierarchical task inheritance allows global defaults with project-specific customizations, reducing configuration duplication while maintaining flexibility.

**Example:**

```yaml
# .moon/tasks/node.yml - Type-based inheritance
tasks:
  install:
    command: 'npm install'
    inputs: [package.json, package-lock.json]

# moon.yml - Project override
tasks:
  install:
    command: 'npm ci --frozen-lockfile'
    inputs: [package.json, package-lock.json]
```

**Techniques:**
- Inheritance order: Global → type → language → tag → project (later overrides)
- Type-based: .moon/tasks/node.yml, .moon/tasks/rust.yml for runtime types
- Language-based: .moon/tasks/javascript.yml, .moon/tasks/typescript.yml
- Tag-based: .moon/tasks/tag-TAGNAME.yml for tag-specific tasks
- Merging: Project tasks merge with inherited tasks
- Complete override: Redefine entire task in moon.yml to fully replace
- Selective override: Define only fields you need to change
