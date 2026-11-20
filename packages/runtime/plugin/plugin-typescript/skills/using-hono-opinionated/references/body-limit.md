# body-limit: Cap Request Payload Size

Apply `bodyLimit` from `hono/body-limit`. Set a low global `maxSize` (in bytes) and raise it per-route for uploads; default response is 413.

```typescript
import {bodyLimit} from "hono/body-limit";

app.use("*", bodyLimit({maxSize: 100 * 1024})); // 100KB global
app.post("/upload", bodyLimit({maxSize: 10 * 1024 * 1024}), async (c) => {
  await c.req.parseBody();
});
```
