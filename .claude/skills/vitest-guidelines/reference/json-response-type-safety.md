# json-response-type-safety: JSON Response Type Safety in Tests

**Guideline:** Define response interfaces and cast JSON results to maintain type safety in tests.

**Rationale:** Type casting prevents unsafe assignment errors and enables IDE autocomplete and refactoring.

**Example:**

```typescript
// Define interfaces at file top
interface User {
  id: string;
  email: string;
}

interface ErrorResponse {
  title: string;
  status: number;
}

// Use in tests
it("should return user", async () => {
  const res = await app.request("/api/users/123");
  const json = (await res.json()) as User;
  expect(json.id).toBe("123"); // ✅ Type-safe
});

it("should return error", async () => {
  const res = await app.request("/api/users/invalid");
  const json = (await res.json()) as ErrorResponse;
  expect(json.status).toBe(404); // ✅ Type-safe
});
```

**Techniques:**
- Create test-specific interfaces at the top of test files
- Use type assertion: `const json = (await res.json()) as ExpectedType`
- Reuse interfaces across multiple tests
- Match interface properties to actual API response shape
- Cast before property access to avoid unsafe-assignment ESLint errors
