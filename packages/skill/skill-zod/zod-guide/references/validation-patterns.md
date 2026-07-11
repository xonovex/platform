# validation-patterns: safeParse vs parse vs pipe

- `.safeParse()` — external input (API, user, services); returns `{ success, data | error }`, never throws.
- `.parse()` — trusted/controlled data only; throws on failure.
- `.pipe()` — feed a `.transform()` result into a second schema for multi-step, type-safe validation.

```typescript
const result = CreateUserSchema.safeParse(req.body);
if (!result.success) {
  return res.status(400).json({error: result.error.issues});
}
const user = await createUser(result.data);
```
