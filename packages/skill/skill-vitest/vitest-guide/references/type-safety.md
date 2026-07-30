# type-safety: Type Safety in Tests

`res.json()` resolves to `any`: cast it to a declared interface before property access, or `@typescript-eslint/no-unsafe-assignment` fires and you lose autocomplete. Define reusable response interfaces at the file top.

```typescript
interface User {
  id: string;
  email: string;
}
interface ErrorResponse {
  title: string;
  status: number;
  issues?: {path: string[]; message: string}[];
}

const json = (await res.json()) as User;
expect(json.email).toBe("test@example.com");

const err = (await res.json()) as ErrorResponse;
expect(err.status).toBe(400);
```
