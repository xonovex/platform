# validation: Zod Validation Pattern

**Guideline:** Validate request params, body, and query at route edges using reusable Zod middleware.

**Rationale:** Early validation prevents invalid data from reaching business logic, provides consistent error responses.

**Example:**

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

**Techniques:**

- Define schemas: Create Zod objects for body, params, query with strict types
- Type inference: Use z.infer<typeof Schema> to export TypeScript types
- safeParse: Use safeParse() instead of parse() to avoid exceptions
- Middleware pattern: Create validateBody/validateParams/validateQuery factories
- Type generics: Accept T extends z.ZodType for reusable middleware
- Error response: Return 400 with result.error.flatten() containing field errors
- Transform: Use .transform(Number) and .pipe() to coerce and validate strings
