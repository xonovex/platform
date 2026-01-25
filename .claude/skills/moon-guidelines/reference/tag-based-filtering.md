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
- #tag:task syntax: Target all projects with specific tag
- Tenant tags: Use #tenant-NAME:task for tenant-specific execution
- Multiple tags: Combine with --query and regex for complex filtering
- Task inheritance: Define .moon/tasks/tag-*.yml for tag-specific tasks
- Project tags: Assign tags in moon.yml for filtering categorization
- Scoped execution: Run across multiple projects without enumeration
- Override behavior: Project-level tasks override tag-inherited tasks
