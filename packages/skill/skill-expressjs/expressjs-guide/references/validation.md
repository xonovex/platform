# validation: Zod Validation Middleware (Express glue)

## Guideline

Validate request params, body, and query at route edges using reusable Zod middleware.

## Rationale

Early validation prevents invalid data from reaching business logic, provides consistent error responses.

Schema design — defining schemas, `z.infer` types, `safeParse` vs `parse`, transforms, refinements, and defaults — belongs to **zod-guide**. This file covers only the Express glue that wires a schema into the request pipeline.

## Example

```typescript
// schemas/users.ts
export const CreateUserSchema = z.object({
  email: z.string().email(),
  name: z.string().min(1).max(100),
  password: z.string().min(8),
});

export const UserParamsSchema = z.object({id: z.string().uuid()});

export type CreateUser = z.infer<typeof CreateUserSchema>;

// middleware/validate.ts
export function validateBody<T extends z.ZodType>(schema: T) {
  return (req: Request, res: Response, next: NextFunction) => {
    const result = schema.safeParse(req.body);
    if (!result.success) {
      return res
        .status(400)
        .json({error: "Validation failed", details: result.error.flatten()});
    }
    req.body = result.data;
    next();
  };
}
```

## Techniques (Express glue — schema design lives in zod-guide)

- Middleware pattern: Create validateBody/validateParams/validateQuery factories
- Type generics: Accept T extends z.ZodType for reusable middleware
- Assign back: Set `req.body = result.data` so handlers receive parsed, coerced input
- Error response: Return 400 with result.error.flatten() containing field errors
