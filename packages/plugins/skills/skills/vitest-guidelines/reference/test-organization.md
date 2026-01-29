# test-organization: Test Directory Structure and Suite Organization

**Guideline:** Organize tests by endpoint/feature with nested describe blocks and clear directory structure.

**Rationale:** Well-organized tests are easy to navigate, maintain, and update.

**Example:**

```
test/
├── app.test.ts           # Application-level tests
├── health.test.ts        # Health check endpoints
└── api/
    ├── users.test.ts     # User CRUD operations
    └── auth.test.ts      # Authentication flows
```

**Techniques:**

- Create `test/` directory at package root
- Mirror API structure in test files (e.g., `test/api/users.test.ts`)
- Use nested `describe` blocks for endpoint grouping
- Name test files after features: `{feature}.test.ts`
- Group related tests in inner `describe` blocks
- Use descriptive `it` statements that explain expected behavior
- Test both success and error cases for each endpoint
