# error-handling: Centralized Error Handler

A single error handler uses the four-parameter signature `(err, req, res, next)` — Express only treats a four-arg function as an error handler. Delegate to `next(err)` when headers have already been sent. Map validation errors and explicitly typed HTTP errors; avoid string matching on `err.name`. Expose internal messages/stacks only outside production, log through a redacting structured logger, and register the handler last.

```typescript
interface HttpError extends Error {
  readonly status: number;
  readonly publicMessage: string;
}

const isHttpError = (error: unknown): error is HttpError =>
  error instanceof Error &&
  "status" in error &&
  typeof error.status === "number" &&
  error.status >= 400 &&
  error.status <= 599 &&
  "publicMessage" in error &&
  typeof error.publicMessage === "string";

export function errorHandler(
  err: unknown,
  req: Request,
  res: Response,
  next: NextFunction,
): void {
  if (res.headersSent) {
    next(err);
    return;
  }

  logger.error({err, method: req.method, path: req.path}, "Request failed");

  if (err instanceof ZodError) {
    res.status(400).json({error: "Validation error", details: err.flatten()});
    return;
  }

  if (isHttpError(err)) {
    res.status(err.status).json({error: err.publicMessage});
    return;
  }

  if (process.env.NODE_ENV !== "production" && err instanceof Error) {
    res.status(500).json({
      error: "Internal server error",
      message: err.message,
      stack: err.stack,
    });
    return;
  }

  res.status(500).json({error: "Internal server error"});
}
```

The logger configuration must redact authorization headers, cookies, request bodies, and secret-bearing error fields. Do not log those values from the middleware itself.
