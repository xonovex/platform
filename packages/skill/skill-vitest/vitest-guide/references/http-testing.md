# http-testing: HTTP Status Assertions

Assert the exact status the middleware sends. Hono `cors()` returns **204** (no body) for an OPTIONS preflight — not 200. Drive routes with `app.request()`; no live server needed.

```typescript
it("handles CORS preflight", async () => {
  const res = await app.request("/api/users", {
    method: "OPTIONS",
    headers: {
      Origin: "https://example.com",
      "Access-Control-Request-Method": "POST",
    },
  });
  expect(res.status).toBe(204);
  expect(res.headers.get("Access-Control-Allow-Origin")).toBe("*");
});
```

Conventions: 201 create, 200 GET-with-body, 204 DELETE/OPTIONS no-body, 400 validation, 404 missing.
