# v4-migration: Zod v4 Migration Guide

**Guideline:** Use Zod v4 standalone validators like `z.uuid()`, `z.email()`, and `z.iso.datetime()` instead of deprecated v3 string refinement methods.

**Rationale:** Zod v4 changed from string refinements to standalone validators that are more efficient, type-safe, and avoid ESLint deprecation warnings.

**Example:**

```typescript
// ✅ Correct (v4) - standalone validators
const UserSchema = z.object({
  id: z.uuid(), // Not z.string().uuid()
  email: z.email(), // Not z.string().email()
  website: z.url(), // Not z.string().url()
  createdAt: z.iso.datetime(), // Not z.string().datetime()
  birthDate: z.iso.date(), // Not z.string().date()
  checkInTime: z.iso.time(), // Not z.string().time()
});

// Enum usage (unchanged)
enum Status {
  Active = "active",
  Inactive = "inactive",
}
const StatusSchema = z.enum(Status);

// String literal unions (unchanged)
const RoleSchema = z.enum(["admin", "user", "guest"]);

// Default values with transformations
const ConfigSchema = z.object({
  port: z.string().transform(Number).pipe(z.number()).default(3000), // ✅ Default matches output type (number)

  timeout: z.string().transform(Number).pipe(z.number()).default(5000), // ✅ Not .default("5000")
});
```

**Techniques:**
- Replace `z.string().uuid()` with `z.uuid()`
- Replace `z.string().email()` with `z.email()`
- Replace `z.string().url()` with `z.url()`
- Replace `z.string().datetime()` with `z.iso.datetime()`
- Replace `z.string().date()` with `z.iso.date()`
- Replace `z.string().time()` with `z.iso.time()`
- Update defaults to match output types with transformations
- Run linter to verify no deprecation warnings
