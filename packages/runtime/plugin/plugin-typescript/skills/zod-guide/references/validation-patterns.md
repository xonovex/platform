# validation-patterns: multi-step validation with pipe

Pipe transformed values into a second schema to validate each stage before applying a default.

- `.pipe()`: feed a `.transform()` result into a second schema for multi-step, type-safe validation. Apply `.transform().pipe()` before `.default()`.

```typescript
const PortSchema = z
  .string()
  .transform((s) => parseInt(s, 10))
  .pipe(z.number().int().min(1024).max(65535))
  .default(3000);
```
