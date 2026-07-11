# timestamp-testing: Avoid Flaky Timestamp Assertions

Rapid operations can complete within the same millisecond, so `updated.updatedAt !== created.updatedAt` is flaky. Prefer asserting existence/format or a before/after range; assert inequality only after an explicit delay.

```typescript
// ✅ existence + format
expect(updated.updatedAt).toBeDefined();
expect(new Date(updated.updatedAt)).toBeInstanceOf(Date);

// ✅ inequality only after a real delay
await new Promise((r) => setTimeout(r, 50));
const updated = await updateUser(created.id);
expect(new Date(updated.updatedAt).getTime()).toBeGreaterThan(
  new Date(created.createdAt).getTime(),
);
```
