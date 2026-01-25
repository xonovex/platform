# validation-patterns: Validation Patterns

**Guideline:** Use `.safeParse()` for external input; `.parse()` for controlled data; `.pipe()` for type-safe transformations.

**Rationale:** Runtime type safety ensures external input never crashes your application.

**Example:**

```typescript
// API input with safeParse
const result = CreateUserSchema.safeParse(req.body);
if (!result.success) {
  return res.status(400).json({error: result.error.issues});
}
const user = await createUser(result.data);

// Query transformation with pipe
const PaginationSchema = z.object({
  page: z.string().regex(/^\d+$/).transform(Number)
    .pipe(z.number().int().min(1)).default(1),
  limit: z.string().regex(/^\d+$/).transform(Number)
    .pipe(z.number().int().min(1).max(100)).default(10),
});

// Controlled data (safe to use parse)
const validated = InternalSchema.parse(data);
```

**Techniques:**
- `.safeParse()`: Use for external input (API, user, services); never throws
- `.parse()`: Use only for controlled data; throws on validation failure
- Check `result.success`: Always verify before accessing `result.data`
- `.transform()`: Apply type conversion before pipe validation
- `.pipe()`: Chain type-safe transformations for multi-step validation
- `.default()`: Provide fallback values for optional fields
- Error handling: Always `.safeParse()` untrusted input, never `.parse()`
- Error context: Preserve error details in API responses
