# migration-v4: Zod v3 → v4 Migration

Replace deprecated v3 string refinements with v4 standalone validators, and native TS enums with `z.enum()`.

| v3 (deprecated)         | v4                 |
| ----------------------- | ------------------ |
| `z.string().uuid()`     | `z.uuid()`         |
| `z.string().email()`    | `z.email()`        |
| `z.string().url()`      | `z.url()`          |
| `z.string().datetime()` | `z.iso.datetime()` |
| `z.string().date()`     | `z.iso.date()`     |
| `z.string().time()`     | `z.iso.time()`     |
| `z.nativeEnum(E)`       | `z.enum(E)`        |

```typescript
const StatusSchema = z.enum(Status); // native TS enum; v3 required z.nativeEnum(Status)
```
