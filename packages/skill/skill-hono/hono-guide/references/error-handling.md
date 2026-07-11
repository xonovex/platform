# error-handling: RFC 7807 Problem Details for Error Responses

Return errors as RFC 7807 Problem Details: required `type`, `title`, `status`; optional `detail`, `instance`; for validation add an `issues[]` array of `{path, message, code}`. Set `Content-Type: application/json`. Wrap common cases in helpers.

```typescript
import type {Context} from "hono";
import type {z} from "zod";

export interface ProblemDetails {
  type: string;
  title: string;
  status: number;
  detail?: string;
  instance?: string;
  issues?: {path: string[]; message: string; code?: string}[];
}

export function badRequest(c: Context, error: z.ZodError): Response {
  const problem: ProblemDetails = {
    type: "about:blank#bad-request",
    title: "Bad Request",
    status: 400,
    detail: "Request validation failed",
    instance: c.req.path,
    issues: error.issues.map((i) => ({
      path: i.path.map(String),
      message: i.message,
      code: i.code,
    })),
  };
  return new Response(JSON.stringify(problem), {
    status: 400,
    headers: {"Content-Type": "application/json"},
  });
}
```
