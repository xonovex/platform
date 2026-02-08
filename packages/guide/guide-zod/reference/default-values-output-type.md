# default-values-output-type: Default Values Must Match Output Type

**Guideline:** When using `.transform()` or `.pipe()`, ensure `.default()` matches the final output type, not the input type.

**Rationale:** Zod v4 requires defaults to match the transformed output type; mismatched types cause TypeScript errors and prevent proper type narrowing.

**Example:**

```typescript
// ✅ Correct - default matches output type (number)
const PaginationSchema = z.object({
  page: z
    .string()
    .regex(/^\d+$/)
    .transform(Number)
    .pipe(z.number().int().min(1))
    .default(1), // Number, matches output type

  limit: z
    .string()
    .regex(/^\d+$/)
    .transform(Number)
    .pipe(z.number().int().min(1).max(100))
    .default(10), // Number, matches output type
});

type Pagination = z.infer<typeof PaginationSchema>;
// Result: {page: number, limit: number}

// ✅ Boolean transformation
const ConfigSchema = z.object({
  debug: z
    .string()
    .transform((val) => val === "true")
    .pipe(z.boolean())
    .default(false), // Boolean, matches output type

  verbose: z
    .string()
    .transform((val) => val === "1")
    .pipe(z.boolean())
    .default(true), // Boolean, matches output type
});

// ✅ Complex transformation
const PortSchema = z
  .string()
  .transform(Number)
  .pipe(z.number().int().min(1024).max(65535))
  .default(3000); // Number, matches output type

type Port = z.infer<typeof PortSchema>;
// Result: number
```

**Techniques:**

- Match defaults to final output type, not input type
- With `.transform(Number)`, use `.default(1)` not `.default("1")`
- With `.transform(Boolean)`, use `.default(true)` not `.default("true")`
- Check TypeScript compilation errors for type mismatches
- Determine output type by tracing through transformation chain
- Use `z.infer<typeof Schema>` to verify final types
- Test with `safeParse()` to catch runtime type issues
