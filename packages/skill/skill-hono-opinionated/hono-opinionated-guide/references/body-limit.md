# body-limit: Cap Request Payload Size

Apply `bodyLimit` from `hono/body-limit` to reject oversized bodies before fully reading them (stream-based, no `Content-Length` needed). Default response is 413 Payload Too Large. Set a low global `maxSize` and raise it per-route for uploads. `maxSize` is in bytes.

```typescript
import {bodyLimit} from "hono/body-limit";

app.use(
  "*",
  bodyLimit({
    maxSize: 100 * 1024, // 100KB default
    onError: (c) =>
      c.json(
        {
          type: "about:blank#payload-too-large",
          title: "Payload Too Large",
          status: 413,
          detail: "Body exceeds 100KB",
        },
        413,
      ),
  }),
);

// Raise the limit for uploads
app.post("/upload", bodyLimit({maxSize: 10 * 1024 * 1024}), async (c) => {
  await c.req.parseBody();
});
```
