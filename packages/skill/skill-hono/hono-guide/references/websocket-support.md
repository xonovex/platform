# websocket-support: WebSocket Server Setup

`createNodeWebSocket({app})` from `@hono/node-ws` returns methods that depend on `this`. Keep the whole object and call `wsHelpers.injectWebSocket(server)` on it — destructuring `injectWebSocket` loses the binding. The `upgradeWebSocket` factory is safe to destructure.

```typescript
import {serve} from "@hono/node-server";
import {createNodeWebSocket} from "@hono/node-ws";
import {Hono} from "hono";
import {createApp} from "./app.js";

const app = createApp();
const wsHelpers = createNodeWebSocket({app}); // keep the object

const server = serve({fetch: app.fetch, port: 3000});
wsHelpers.injectWebSocket(server); // method call preserves `this`

const {upgradeWebSocket} = wsHelpers; // factory is safe to destructure

export const wsRouter = new Hono();
wsRouter.get(
  "/chat",
  upgradeWebSocket(() => ({
    onMessage(event, ws) {
      const data = JSON.parse(String(event.data)) as {message: string};
      ws.send(JSON.stringify({echo: data.message}));
    },
  })),
);
```
