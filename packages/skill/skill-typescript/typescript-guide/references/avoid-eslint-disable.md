# avoid-eslint-disable: Fix Root Causes, Never Suppress

Never suppress with `eslint-disable`, `@ts-ignore`, or an `any` cast: refactor the code to satisfy the type instead.

```typescript
// ❌ suppress the error
// eslint-disable-next-line @typescript-eslint/no-explicit-any
itemsRouter.openapi(listItemsRoute, itemsController.listItems as any);

// ✅ fix the type mismatch
itemsRouter.openapi(listItemsRoute, (c) => {
  const {page, limit} = c.req.valid("query");
  const {items, total} = itemsService.listItems(page, limit);
  return c.json({items, pagination: {page, limit, total}}, 200);
});
```
