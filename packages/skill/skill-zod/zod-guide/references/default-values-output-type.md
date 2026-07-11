# default-values-output-type: Default Values Must Match Output Type

With `.transform()`/`.pipe()`, `.default(x)` must match the final **output** type, not the input string — Zod v4 type-checks the default against the piped output. Use `.default(1)` after `.transform(Number)`, `.default(false)` after a boolean transform.

```typescript
const PortSchema = z
  .string()
  .transform(Number)
  .pipe(z.number().int().min(1024).max(65535))
  .default(3000); // number, not "3000"

type Port = z.infer<typeof PortSchema>; // number
```
