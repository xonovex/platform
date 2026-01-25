# project-constraints: Enforce Architectural Boundaries

**Guideline:** Define tag-based constraints in workspace to enforce dependency rules between projects.

**Rationale:** Tag-based constraints prevent architectural violations by restricting which project tags can depend on others, enforcing clear module boundaries in monorepos.

**Example:**

```yaml
# .moon/workspace.yml
constraints:
  tagRelationships:
    frontend: [frontend, shared]
    backend: [backend, shared]
    tenant-example: [tenant-example, shared]
```

**Techniques:**
- tagRelationships: Define which tags can depend on which tags
- Validation: Constraints checked during project graph creation and task execution
- Violations: Prevent execution if dependencies violate constraints
- Frontend tags: Can depend on frontend and shared tags only
- Tenant isolation: Each tenant can depend only on own tenant and shared
- moon check: Validate constraints without running tasks
- Tag refactoring: Update moon.yml tags to align with constraints
