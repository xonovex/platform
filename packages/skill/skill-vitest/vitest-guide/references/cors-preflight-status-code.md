# cors-preflight-status-code: CORS Preflight Status Code Expectations

**Guideline:** OPTIONS preflight requests should return HTTP 204 (No Content), not 200 (OK).

**Rationale:** HTTP 204 is the correct status code for OPTIONS requests with no response body.

**Example:**

```typescript
it("should handle CORS preflight", async () => {
  const res = await app.request("/api/users", {
    method: "OPTIONS",
    headers: {
      Origin: "https://example.com",
      "Access-Control-Request-Method": "POST",
      "Access-Control-Request-Headers": "Content-Type",
    },
  });

  expect(res.status).toBe(204); // ✅ Correct
  expect(res.headers.get("Access-Control-Allow-Origin")).toBe("*");
  expect(res.headers.get("Access-Control-Allow-Methods")).toContain("POST");
  expect(res.headers.get("Access-Control-Allow-Headers")).toContain(
    "Content-Type",
  );
});
```

**Techniques:**

- Always expect 204 status for OPTIONS requests in tests
- Verify CORS headers are present (Access-Control-Allow-Origin, etc.)
- Don't expect a response body from OPTIONS
- Use the same pattern for all preflight request tests
- Test both missing and present CORS headers
