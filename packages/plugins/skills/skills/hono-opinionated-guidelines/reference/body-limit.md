# body-limit: Preventing Large Payload Attacks

**Guideline:** Use Body Limit middleware to prevent DoS attacks from oversized request payloads, configuring appropriate limits per endpoint based on expected data.

**Rationale:** Large request bodies can exhaust server memory, cause denial of service, and overwhelm downstream services. Body Limit middleware rejects requests before fully reading them, uses stream reading without Content-Length requirements, and returns 413 Payload Too Large status while remaining configurable per route.

**Example:**

```typescript
import {bodyLimit} from "hono/body-limit";

// Global limit for most endpoints
app.use(
  "*",
  bodyLimit({
    maxSize: 100 * 1024, // 100KB default
    onError: (c) => {
      return c.json(
        {
          type: "about:blank#payload-too-large",
          title: "Payload Too Large",
          status: 413,
          detail: "Request body exceeds 100KB limit",
        },
        413,
      );
    },
  }),
);

// Higher limit for file uploads
app.post(
  "/upload",
  bodyLimit({
    maxSize: 10 * 1024 * 1024, // 10MB for uploads
  }),
  async (c) => {
    const body = await c.req.parseBody();
    // Handle file upload
  },
);

// Stricter limit for JSON APIs
app.use(
  "/api/*",
  bodyLimit({
    maxSize: 50 * 1024, // 50KB for JSON
  }),
);
```

**Techniques:**
- Import `bodyLimit` from `hono/body-limit` and set appropriate `maxSize` in bytes
- Apply globally with app.use() or per-route for flexibility
- Configure different limits for different endpoints (100KB default, 10MB for uploads)
- Provide custom `onError` handler for formatted error responses
- Consider internal vs public services when setting limits
