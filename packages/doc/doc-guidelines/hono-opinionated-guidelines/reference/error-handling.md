# error-handling: RFC 7807 Problem Details for Consistent Error Responses

**Guideline:** Use RFC 7807 Problem Details format for all error responses to provide consistent, structured, machine-readable error information across your API.

**Rationale:** RFC 7807 provides a standard structure recognized across HTTP APIs, enabling clients to parse errors consistently with machine-readable field-level validation details and request context. Built-in extensibility and separation between human and machine-readable fields improves debugging and error handling.

**Example:**

```typescript
import type {Context} from "hono";
import type {z} from "zod";

export interface ProblemDetails {
  type: string;
  title: string;
  status: number;
  detail?: string;
  instance?: string;
  issues?: {
    path: string[];
    message: string;
    code?: string;
  }[];
}

export function badRequest(c: Context, error: z.ZodError): Response {
  const problem: ProblemDetails = {
    type: "about:blank#bad-request",
    title: "Bad Request",
    status: 400,
    detail: "Request validation failed",
    instance: c.req.path,
    issues: error.issues.map((issue) => ({
      path: issue.path.map(String),
      message: issue.message,
      code: issue.code,
    })),
  };

  return new Response(JSON.stringify(problem), {
    status: 400,
    headers: {"Content-Type": "application/json"},
  });
}

export function notFound(c: Context, detail: string): Response {
  const problem: ProblemDetails = {
    type: "about:blank#not-found",
    title: "Not Found",
    status: 404,
    detail,
    instance: c.req.path,
  };

  return new Response(JSON.stringify(problem), {
    status: 404,
    headers: {"Content-Type": "application/json"},
  });
}
```

**Techniques:**

- Define `ProblemDetails` TypeScript interface matching RFC 7807 spec
- Include required fields: type, title, status
- Include optional fields: detail, instance, custom extensions
- For validation errors add `issues` array with field-level details
- Create helper functions for common error types (badRequest, notFound, etc.)
- Return proper HTTP status codes matching the status field
- Set `Content-Type: application/json` header
