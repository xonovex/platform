# timestamp-testing: Avoiding Flaky Timestamp Comparisons

**Guideline:** Don't compare timestamps from rapid operations; they may complete in the same millisecond.

**Rationale:** Fast operations can complete within milliseconds, making timestamp comparisons flaky and unreliable.

**Example:**

```typescript
// ✅ Preferred - verify existence and type
it("should update timestamp", async () => {
  const created = await createUser();
  const updated = await updateUser(created.id);

  expect(updated.updatedAt).toBeDefined();
  expect(typeof updated.updatedAt).toBe("string");
  expect(new Date(updated.updatedAt)).toBeInstanceOf(Date);
});

// ✅ Alternative - add explicit delay
it("should update timestamp with delay", async () => {
  const created = await createUser();
  await new Promise((resolve) => setTimeout(resolve, 10)); // 10ms delay
  const updated = await updateUser(created.id);

  expect(updated.updatedAt).not.toBe(created.updatedAt);
});

// ✅ Good - verify timestamp is after a known point
it("should have recent timestamp", async () => {
  const before = new Date().toISOString();
  const created = await createUser();

  expect(created.createdAt >= before).toBe(true);
});

// ✅ Good - verify timestamp changed (with reasonable buffer)
it("should update timestamp", async () => {
  const created = await createUser();
  const createdTime = new Date(created.createdAt).getTime();

  await new Promise((resolve) => setTimeout(resolve, 50));
  const updated = await updateUser(created.id);
  const updatedTime = new Date(updated.updatedAt).getTime();

  expect(updatedTime).toBeGreaterThan(createdTime);
});
```

**Techniques:**
- Verify timestamp exists and is correct type (preferred approach)
- Add explicit delay (10-50ms) before second operation if comparison needed
- Use timestamp ranges (before/after) instead of exact equality
- Verify timestamp format/structure instead of comparing values
- Mock time if deterministic testing is required
