# responses: Response Patterns and Status Codes

Wrap every response in a `{data}` or `{error, details?}` envelope. Status codes: 200 GET/PUT/PATCH, 201 POST (add `Location` header), 204 DELETE (`res.status(204).send()`, no body), 400 validation (`{error, details}`), 401/403 auth, 404/5xx via the error handler.

```typescript
res.json({data: result}); // 200
res.json({data: items, pagination: {page: 1, limit: 20, total: 100}});
res.status(400).json({error: "Validation failed", details: {...}});
res.status(204).send(); // no content
```
