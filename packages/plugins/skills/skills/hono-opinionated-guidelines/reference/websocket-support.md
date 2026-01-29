# websocket-support: WebSocket Server Setup and Route Patterns

**Guideline:** Keep the object reference returned from `createNodeWebSocket()` and call methods on that object rather than destructuring, to maintain proper `this` binding.

**Rationale:** JavaScript methods that rely on `this` context lose their binding when destructured. The `@hono/node-ws` package's `createNodeWebSocket()` returns methods that depend on `this` to access internal state. Keeping the object reference preserves the binding while factory functions like `upgradeWebSocket` can safely be destructured.

**Example:**

```typescript
import {serve} from "@hono/node-server";
import {createNodeWebSocket} from "@hono/node-ws";
import {Hono} from "hono";
import {createApp} from "./app.js";

const app = createApp();

// ✅ CORRECT - Keep object reference
const wsHelpers = createNodeWebSocket({app});

const server = serve({
  fetch: app.fetch,
  port: 3000,
});

// Call method on object to maintain binding
wsHelpers.injectWebSocket(server);

// Destructuring upgradeWebSocket is safe (it's a factory)
const {upgradeWebSocket} = createNodeWebSocket({app: new Hono()});

export const wsRouter = new Hono();

wsRouter.get(
  "/chat",
  upgradeWebSocket(() => ({
    onOpen(_evt, ws) {
      console.log("Client connected");
    },
    onMessage(event, ws) {
      const data = JSON.parse(String(event.data)) as {message: string};
      ws.send(JSON.stringify({echo: data.message}));
    },
    onClose(_evt, ws) {
      console.log("Client disconnected");
    },
    onError(evt, ws) {
      console.error("WebSocket error:", evt);
    },
  })),
);
```

**Techniques:**

- Import `createNodeWebSocket` from `@hono/node-ws` and call with `{app}` parameter
- Keep entire object reference (e.g., `wsHelpers`) for method-based APIs
- Call methods on the object: `wsHelpers.injectWebSocket(server)`
- Destructuring factory functions like `upgradeWebSocket` is safe
- Define WebSocket handlers with onOpen, onMessage, onClose, onError callbacks
- Use TypeScript types for message data to ensure type safety
