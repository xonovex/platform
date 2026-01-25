# schema-organization: Schema Organization Patterns

**Guideline:** Name schemas with PascalCase suffix, compose with `.merge()`, `.pick()`, `.partial()`, and always infer types using `z.infer<typeof Schema>`.

**Rationale:** Schema-first design provides single source of truth for types and validation, enabling reusable composition and reducing duplication.

**Example:**

```typescript
// ✅ Base schema with common fields
const BaseEntitySchema = z.object({
  id: z.uuid(),
  createdAt: z.iso.datetime(),
  updatedAt: z.iso.datetime(),
});

// ✅ Extend with .merge()
export const UserSchema = BaseEntitySchema.merge(
  z.object({
    email: z.email(),
    name: z.string().min(1).max(100),
    role: z.enum(["admin", "user", "guest"]),
  }),
);

// ✅ Infer type from schema
export type User = z.infer<typeof UserSchema>;

// ✅ Create input schemas with .pick()
export const CreateUserSchema = UserSchema.pick({
  email: true,
  name: true,
  role: true,
});

export type CreateUser = z.infer<typeof CreateUserSchema>;

// ✅ Make fields optional with .partial()
export const UpdateUserSchema = UserSchema.pick({
  email: true,
  name: true,
  role: true,
}).partial();

export type UpdateUser = z.infer<typeof UpdateUserSchema>;

// ✅ Simplified view with .pick()
export const UserSummarySchema = UserSchema.pick({
  id: true,
  email: true,
  name: true,
});

export type UserSummary = z.infer<typeof UserSummarySchema>;

// Usage in API
async function createUser(input: unknown): Promise<User> {
  const data = CreateUserSchema.parse(input);
  // data is type-safe: CreateUser
  const user = await db.users.create(data);
  return UserSchema.parse(user);
}
```

**Techniques:**
- Name all schemas with PascalCase + "Schema" suffix
- Define schemas at module level for reuse
- Always infer types: `type User = z.infer<typeof UserSchema>`
- Use `.merge()` to combine schemas
- Use `.pick()` to select specific fields
- Use `.partial()` to make fields optional
- Create base schemas for common patterns
- Keep schemas close to their usage
