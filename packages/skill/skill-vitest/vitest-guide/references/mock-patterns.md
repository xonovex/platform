# mock-patterns: Mock Typing

`vi.fn()` already returns a `Mock` — call `.mockResolvedValue` / `.mockRejectedValue` / `.mockReturnValue` directly, no cast or generic. Add `vi.fn<...>()` type params only when an assertion genuinely needs them; otherwise let TypeScript infer.

```typescript
const getUserFn = vi.fn();
getUserFn.mockResolvedValue({id: "123", email: "test@example.com"});

const result = await getUserFn("123");
expect(getUserFn).toHaveBeenCalledWith("123");

getUserFn.mockRejectedValue(new Error("Not found"));
await expect(getUserFn("999")).rejects.toThrow("Not found");
```
