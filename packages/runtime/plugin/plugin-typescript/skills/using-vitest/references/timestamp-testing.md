# timestamp-testing: Avoid Flaky Timestamp Assertions

Rapid operations can share a millisecond, so `updated.updatedAt !== created.updatedAt` is flaky. Assert existence/format, or put the delay in the **test body** (not the implementation under test) before comparing.

```typescript
// existence + format
expect(updated.updatedAt).toBeDefined();

// inequality only after a delay in the test itself
await new Promise((r) => setTimeout(r, 50));
const updated = await updateUser(created.id);
expect(new Date(updated.updatedAt).getTime()).toBeGreaterThan(
  new Date(created.createdAt).getTime(),
);
```
