# error-handling: Centralized Error Handler

**Guideline:** Implement single error handler middleware that handles Zod errors, custom errors, and hides stack traces in production.

**Rationale:** Centralized error handling ensures consistent error responses and prevents information leakage.

**Example:**

```typescript
export function errorHandler(
  err: unknown,
  req: Request,
  res: Response,
  next: NextFunction,
): void {
  console.error("Error:", err);

  if (err instanceof ZodError) {
    res.status(400).json({error: "Validation error", details: err.flatten()});
    return;
  }

  if (err instanceof Error) {
    if (err.name === "NotFoundError") {
      res.status(404).json({error: err.message});
      return;
    }
    if (err.name === "UnauthorizedError") {
      res.status(401).json({error: err.message});
      return;
    }
    if (err.name === "ForbiddenError") {
      res.status(403).json({error: err.message});
      return;
    }

    if (process.env.NODE_ENV !== "production") {
      res.status(500).json({
        error: "Internal server error",
        message: err.message,
        stack: err.stack,
      });
      return;
    }
  }

  res.status(500).json({error: "Internal server error"});
}
```

**Techniques:**

- Four parameters: (err, req, res, next) required for Express to recognize error handler
- Zod handling: Check instanceof ZodError, return 400 with details
- Custom errors: Match error.name against NotFoundError, UnauthorizedError, ForbiddenError
- Development mode: Show error.message and error.stack when NODE_ENV !== "production"
- Production mode: Hide details, return generic "Internal server error" message
- Register last: Error handler must be last middleware in app.use() chain
- Logging: Log errors to console for debugging and monitoring
