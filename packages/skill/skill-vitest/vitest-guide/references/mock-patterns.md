# mock-patterns: Simple Type Casting for Vitest Mocks

**Guideline:** Use simple type casting for Vitest mocks instead of complex generic typing.

**Rationale:** Simple casting is more maintainable and readable than complex generic typing.

**Example:**

```typescript
import {expect, it, vi} from "vitest";

interface User {
  id: string;
  email: string;
}

// ✅ Simple - easy to read and maintain
it("should call user service", async () => {
  const mockUser: User = {id: "123", email: "test@example.com"};
  const getUserFn = vi.fn();
  getUserFn.mockResolvedValue(mockUser);

  const result = await getUserFn("123");
  expect(result).toEqual(mockUser);
  expect(getUserFn).toHaveBeenCalledWith("123");
});

// ✅ Good - reject without a redundant cast (vi.fn() already returns a Mock)
it("should handle error", async () => {
  const getUserFn = vi.fn();
  getUserFn.mockRejectedValue(new Error("Not found"));

  await expect(getUserFn("999")).rejects.toThrow("Not found");
});

// ✅ Good - simple sync mock
it("should transform data", () => {
  const transformFn = vi.fn();
  transformFn.mockReturnValue({transformed: true});

  const result = transformFn({input: "data"});
  expect(result.transformed).toBe(true);
});
```

**Techniques:**

- Create mocks with `vi.fn()` without complex generics
- `vi.fn()` already returns a `Mock`, so call `.mockResolvedValue` / `.mockRejectedValue` directly without casting; only type the variable at declaration (`vi.fn<...>()`) when a type is genuinely needed
- Let TypeScript infer types from usage
- Only add types when necessary for test assertions
- Prefer simple casting over advanced mock type patterns
