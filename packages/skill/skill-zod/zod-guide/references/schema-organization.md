# schema-organization: Schema Composition

Name schemas `PascalCase` + `Schema`, define at module level near usage, and derive variants instead of redefining. Infer every type with `z.infer<typeof Schema>`.

- `.extend({ ... })` — add fields; `.extend(Other.shape)` merges two schemas (v4 replacement for deprecated `.merge()`)
- `.pick({ field: true })` / `.omit({ field: true })` — select or drop fields
- `.partial()` — make fields optional (e.g. update DTOs)

```typescript
const BaseEntitySchema = z.object({
  id: z.uuid(),
  createdAt: z.iso.datetime(),
});

export const UserSchema = BaseEntitySchema.extend({
  email: z.email(),
  name: z.string().min(1).max(100),
  role: z.enum(["admin", "user", "guest"]),
});
export type User = z.infer<typeof UserSchema>;

export const CreateUserSchema = UserSchema.pick({
  email: true,
  name: true,
  role: true,
});
export const UpdateUserSchema = CreateUserSchema.partial();
export type CreateUser = z.infer<typeof CreateUserSchema>;
```
