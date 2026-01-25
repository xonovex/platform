# project-references-path-resolution: Path Resolution in TypeScript Project References

**Guideline:** Verify directory structure before configuring project references; calculate relative paths carefully from current package to target.

**Rationale:** Incorrect paths cause "File not found" errors and build failures in CI/CD pipelines.

**Example:**

```bash
# Verify structure
$ ls packages/
templates/ shared/ apps/

# Calculate path: packages/templates/my-template → packages/shared/shared-core
# Up 2 levels (..), down through shared/, then shared-core/
```

```json
// Correct reference
{"references": [{"path": "../../shared/shared-core"}]}

// Wrong - goes up 3 levels
// {"path": "../../../shared/shared-core"}  ❌

// Wrong - missing "shared" directory
// {"path": "../../shared-core"}  ❌
```

**Techniques:**
- ls verification: Use `ls` to check actual directory structure before configuring
- Relative paths: Calculate from current package up (..) then down to target
- tsc --build: Test with TypeScript build to verify paths resolve
- Error messages: Read build errors to identify which paths failed
- Monorepo structure: Understand packages/, apps/, shared/ hierarchy
- Counting levels: Each up (..) represents one directory level
- Off-by-one errors: Most common mistake - verify counting twice
