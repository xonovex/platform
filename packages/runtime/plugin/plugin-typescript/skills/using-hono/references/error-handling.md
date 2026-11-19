# error-handling: RFC 9457 Problem Details for Error Responses

Return errors as RFC 9457 Problem Details with `Content-Type: application/problem+json`. This API's stable problem shape includes the canonical `type`, `title`, and `status` members plus optional `detail` and `instance`; validation problems add an `issues[]` extension of `{path, message, code}`. Use a stable problem-type URI controlled by the API for custom problem types. Wrap common cases in helpers.

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
    type: "https://api.example.com/problems/validation-error",
    title: "Request validation failed",
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
    headers: {"Content-Type": "application/problem+json"},
  });
}
```
