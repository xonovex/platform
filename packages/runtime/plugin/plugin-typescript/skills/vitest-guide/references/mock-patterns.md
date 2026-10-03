# mock-patterns: Mock Typing

Test async rejection with `.mockRejectedValue`, then assert with `.rejects.toThrow`.

```typescript
const getUserFn = vi.fn();
getUserFn.mockRejectedValue(new Error("Not found"));
await expect(getUserFn("999")).rejects.toThrow("Not found");
```
