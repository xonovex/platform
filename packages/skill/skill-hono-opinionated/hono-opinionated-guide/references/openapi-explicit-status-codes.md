# openapi-explicit-status-codes: Pass Explicit Status to Every c.json()

Every `c.json()` in an OpenAPI handler must pass its status code explicitly — the response type is a discriminated union keyed on status, and TypeScript can't narrow it without the literal. Match the code to a key in the route's `responses` object. Never rely on the implicit 200.

```typescript
itemsRouter.openapi(getItemRoute, (c) => {
  const item = itemsService.findById(c.req.valid("param").id);
  if (!item) {
    return c.json(
      {
        type: "about:blank#not-found",
        title: "Not Found",
        status: 404,
        detail: "Item not found",
      },
      404,
    );
  }
  return c.json(item, 200);
});
```
