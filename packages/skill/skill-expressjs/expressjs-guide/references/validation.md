# validation: Zod Validation Middleware (Express glue)

Wire schemas at route edges with reusable `validateBody`/`validateParams`/`validateQuery` factories: `safeParse` the segment, return 400 with `error.flatten()` on failure, and assign the parsed result back (`req.body = result.data`) so handlers receive coerced input. Schema design (`z.infer`, transforms, refinements, defaults) belongs to **zod-guide**.

```typescript
export function validateBody<T extends z.ZodType>(schema: T) {
  return (req: Request, res: Response, next: NextFunction) => {
    const result = schema.safeParse(req.body);
    if (!result.success) {
      return res
        .status(400)
        .json({error: "Validation failed", details: result.error.flatten()});
    }
    req.body = result.data; // parsed, coerced input
    next();
  };
}
```
