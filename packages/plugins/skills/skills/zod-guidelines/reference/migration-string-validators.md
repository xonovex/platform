# string-validation-method-changes: String Validation Method Changes

**Guideline:** Use standalone validators `z.uuid()` and `z.iso.datetime()` instead of deprecated string refinements `z.string().uuid()` and `z.string().datetime()`.

**Rationale:** Zod v4 deprecated string refinements; standalone methods are more efficient and prevent ESLint deprecation warnings.

**Example:**

```typescript
// ✅ Correct - Zod v4 standalone validators
const EventSchema = z.object({
  id: z.uuid(), // Standalone UUID validator
  eventId: z.uuid(), // Use for any UUID field
  scheduledAt: z.iso.datetime(), // Standalone datetime validator
  createdAt: z.iso.datetime(), // Use for any ISO datetime
  updatedAt: z.iso.datetime(),
});

// Infer type
type Event = z.infer<typeof EventSchema>;
// Result: {id: string, eventId: string, scheduledAt: string, ...}

// Validation
const result = EventSchema.safeParse(data);
if (result.success) {
  console.log(result.data.id); // Type-safe access
}
```

**Techniques:**

- Replace `z.string().uuid()` with `z.uuid()`
- Replace `z.string().email()` with `z.email()`
- Replace `z.string().url()` with `z.url()`
- Replace `z.string().datetime()` with `z.iso.datetime()`
- Replace `z.string().date()` with `z.iso.date()`
- Replace `z.string().time()` with `z.iso.time()`
- Run linter to verify no deprecation warnings
- Test validation to ensure behavior unchanged
