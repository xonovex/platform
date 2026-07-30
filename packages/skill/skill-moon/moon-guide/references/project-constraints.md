# project-constraints: Enforce Architectural Boundaries

Define `constraints` in `.moon/workspace.yml` to restrict which project tags/layers may depend on which. Constraints are checked during project-graph creation and violations block task execution; run `moon check` to validate without running tasks.

```yaml
# .moon/workspace.yml
constraints:
  tagRelationships:
    # source tag: allowed dependency tags: a project carrying the source tag
    # may only depend on projects whose tags appear in the list
    frontend: [frontend, shared]
    backend: [backend, shared]
    cli: [cli, shared, backend]
  enforceLayerRelationships: true # Moon 2.0
```

`enforceLayerRelationships` forbids a lower layer depending on a higher one. Hierarchy (lowest → highest):

1. `configuration` may depend on: nothing
2. `library` may depend on: configuration, library
3. `application` may depend on: configuration, library, application
