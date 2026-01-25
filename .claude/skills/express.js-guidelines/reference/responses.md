# responses: Response Patterns and Status Codes

**Guideline:** Use consistent JSON response shapes with appropriate HTTP status codes.

**Rationale:** Standardized responses simplify client-side handling and follow HTTP semantics.

**Example:**

```typescript
// Success
res.json({data: result});

// Success with pagination
res.json({data: items, pagination: {page: 1, limit: 20, total: 100}});

// Error
res.status(400).json({error: "Validation failed", details: {...}});

// No content (deletion)
res.status(204).send();
```

**Techniques:**
- Data wrapper: Wrap all responses in {data: ...} or {error: ...} envelope
- Status 200: GET, PUT, PATCH success - return {data: modified}
- Status 201: POST success - return {data: created} with Location header
- Status 204: DELETE success - return no content
- Status 400: Validation errors - return {error: "message", details: {...}}
- Status 401/403: Auth errors - return {error: "message"}
- Status 404/5xx: Return {error: message} via error handler
