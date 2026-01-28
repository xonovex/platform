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
  enforceLayerRelationships: true
```

**Techniques:**

## tagRelationships
- Define which tags can depend on which tags
- Key: Source tag, Value: Allowed dependency tags
- Projects with source tag can only depend on projects with allowed tags

```yaml
constraints:
  tagRelationships:
    # Frontend can depend on frontend and shared
    frontend: [frontend, shared]
    # Backend can depend on backend and shared
    backend: [backend, shared]
    # CLI can depend on cli, shared, and backend
    cli: [cli, shared, backend]
```

## enforceLayerRelationships (Moon 2.0)
- Enforce that libraries cannot depend on applications
- Prevents circular dependencies between layers

```yaml
constraints:
  enforceLayerRelationships: true
```

Layer hierarchy (lowest to highest):
1. `configuration` - Can depend on: nothing
2. `library` - Can depend on: configuration, library
3. `application` - Can depend on: configuration, library, application

## Validation
- Constraints checked during project graph creation
- Violations prevent task execution
- Run `moon check` to validate without running tasks

## Common Patterns

**Tenant isolation:**
```yaml
tagRelationships:
  tenant-acme: [tenant-acme, shared]
  tenant-beta: [tenant-beta, shared]
```

**Feature boundaries:**
```yaml
tagRelationships:
  auth: [auth, shared, database]
  api: [api, shared, auth, database]
  web: [web, shared, api]
```

## Troubleshooting
- Check project tags in `moon.yml`
- Verify `dependsOn` doesn't violate constraints
- Use `moon query projects` to inspect project relationships
