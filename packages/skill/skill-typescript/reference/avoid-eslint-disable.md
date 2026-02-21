# avoid-eslint-disable: Never Suppress Warnings - Fix Root Causes

**Guideline:** Never use `eslint-disable`, `@ts-ignore`, or `any` types; fix root causes instead.

**Rationale:** Suppressing warnings hides architectural issues and creates technical debt.

**Example:**

```typescript
// ❌ WRONG: Suppress error
// eslint-disable-next-line @typescript-eslint/no-explicit-any
itemsRouter.openapi(listItemsRoute, itemsController.listItems as any);

// ✅ CORRECT: Fix type mismatch
itemsRouter.openapi(listItemsRoute, (c) => {
  const {page, limit} = c.req.valid("query");
  const {items, total} = itemsService.listItems(page, limit);
  return c.json({items, pagination: {page, limit, total}}, 200);
});
```

**Techniques:**

- Read error message carefully and understand what it indicates
- Research framework documentation and TypeScript behavior
- Refactor code to align with type expectations
- Never use `any`, `@ts-ignore`, or `eslint-disable` comments
- Verify fix resolves issue completely with type checking
- Run tests after refactoring to ensure correctness
