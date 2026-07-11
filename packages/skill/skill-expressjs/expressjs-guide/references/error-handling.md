# error-handling: Centralized Error Handler

A single error handler with the four-parameter signature `(err, req, res, next)` — Express only treats a four-arg function as an error handler. Branch `ZodError` → 400 with `err.flatten()`, match custom `err.name` (`NotFoundError`/`UnauthorizedError`/`ForbiddenError`) to 404/401/403, and expose `err.message`/`err.stack` only when `NODE_ENV !== "production"`; otherwise return a generic 500. Register it last.

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
