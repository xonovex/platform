# code-align: Analyze Code Alignment Between Similar Implementations

**Guideline:** Compare implementations to identify structural differences, inconsistencies, and alignment opportunities.

**Rationale:** Similar code should follow consistent patterns. Identifies bugs from missing error handling, outdated APIs, or logic differences.

**Example:**

```
// Compare two API implementations
File A: imports v3.5, uses async/await, has try-catch
File B: imports v2.1, uses callbacks, missing error handling

Report: Version mismatch (v3.5 vs v2.1), pattern divergence (async vs callback)
```

**Techniques:**

- Load both files and detect code type (TypeScript, Python, Go, etc.)
- Compare imports, dependencies, and version constraints across implementations
- Analyze interfaces, types, and function signatures for structural differences
- Examine error handling patterns, validation logic, and edge cases
- Review code organization, constants, and configuration defaults
- Generate alignment report with percentage match and critical differences
- Document recommendations for standardizing diverged implementations
