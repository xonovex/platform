# http-testing: HTTP Testing Patterns

**Guideline:** Use correct HTTP status codes in test assertions (204 for OPTIONS/DELETE, 201 for creation, 200 for GET).

**Rationale:** Correct status codes ensure tests align with HTTP standards and detect real failures.

**Example:**

```typescript
// CORS Preflight
it("should handle CORS preflight", async () => {
  const res = await app.request("/api/users", {
    method: "OPTIONS",
    headers: {
      Origin: "https://example.com",
      "Access-Control-Request-Method": "POST",
    },
  });

  expect(res.status).toBe(204); // Hono cors() returns 204 for preflight; the CORS spec allows any 2xx, so assert what the configured middleware sends
  expect(res.headers.get("Access-Control-Allow-Origin")).toBeDefined();
});

// Status Code Assertions
it("should create resource", async () => {
  const res = await app.request("/api/users", {method: "POST", body: data});
  expect(res.status).toBe(201); // ✅ Created
});

it("should get resource", async () => {
  const res = await app.request("/api/users/123");
  expect(res.status).toBe(200); // ✅ OK with body
});

it("should delete resource", async () => {
  const res = await app.request("/api/users/123", {method: "DELETE"});
  expect(res.status).toBe(204); // ✅ No Content
});

it("should validate input", async () => {
  const res = await app.request("/api/users", {method: "POST", body: invalid});
  expect(res.status).toBe(400); // ✅ Bad Request
});

it("should handle missing resource", async () => {
  const res = await app.request("/api/users/999");
  expect(res.status).toBe(404); // ✅ Not Found
});
```

**Techniques:**

- Assert the preflight (OPTIONS) status your CORS middleware actually returns (Hono `cors()` returns 204 with no body); the CORS spec permits any 2xx, so don't treat 204 as an absolute requirement
- Use 204 for successful DELETE operations without response body
- Use 201 for resource creation (POST that creates new resource)
- Use 200 for successful operations with response body
- Use 400 for validation failures and bad requests
- Use 404 for missing resources
