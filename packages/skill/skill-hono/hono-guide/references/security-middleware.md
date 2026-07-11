# security-middleware: Built-in Security Middleware

Use Hono's built-in middleware instead of external deps: `secureHeaders`, `csrf`, `basicAuth`, `bearerAuth`, `ipRestriction` (all from `hono/*` subpaths). Apply `secureHeaders()` first so it covers every route. `ipRestriction` takes `allowList`/`denyList` in CIDR notation.

```typescript
import {basicAuth} from "hono/basic-auth";
import {bearerAuth} from "hono/bearer-auth";
import {csrf} from "hono/csrf";
import {ipRestriction} from "hono/ip-restriction";
import {secureHeaders} from "hono/secure-headers";

app.use("*", secureHeaders());
app.use("/forms/*", csrf());
app.use(
  "/api/*",
  bearerAuth({verifyToken: async (token, c) => tokenService.verify(token)}),
);
app.use(
  "/internal/*",
  ipRestriction({denyList: [], allowList: ["192.168.0.0/16", "10.0.0.0/8"]}),
);
```
