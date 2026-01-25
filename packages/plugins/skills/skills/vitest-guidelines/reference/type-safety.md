# type-safety: Type Safety in Tests

**Guideline:** Always define interfaces for response shapes and cast JSON parsing results to maintain type safety.

**Rationale:** Type assertions enable IDE autocomplete and prevent unsafe assignment errors.

**Example:**

```typescript
import {describe, expect, it} from "vitest";

// Define reusable interfaces at the top
interface User {
  id: string;
  email: string;
  name: string;
  createdAt: string;
}

interface ErrorResponse {
  title: string;
  status: number;
  issues?: {path: string[]; message: string}[];
}

describe("Users API", () => {
  it("should create user", async () => {
    const res = await app.request("/api/users", {
      method: "POST",
      body: JSON.stringify({email: "test@example.com"}),
    });

    const json = (await res.json()) as User;
    expect(json.email).toBe("test@example.com"); // Type-safe!
  });

  it("should validate email", async () => {
    const res = await app.request("/api/users", {
      method: "POST",
      body: JSON.stringify({email: "invalid"}),
    });

    const json = (await res.json()) as ErrorResponse;
    expect(json.status).toBe(400);
    expect(json.issues).toBeDefined(); // Type-safe!
  });
});
```

**Techniques:**
- Define interfaces for expected response shapes at the top of test files
- Cast all JSON parsing results: `const json = (await res.json()) as ResponseType`
- Create reusable interfaces for common response types (User, ErrorResponse, etc.)
- Use the same interfaces across multiple tests in the same file
- Match interface properties to actual API response shapes
