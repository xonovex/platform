# schema-organization: Naming and Placement

Name schemas `PascalCase` + `Schema` (e.g. `UserSchema`), and define them at module level near their usage so they can be reused across the module.

```typescript
export const UserSchema = z.object({
  id: z.uuid(),
  email: z.email(),
  name: z.string().min(1).max(100),
});
```
