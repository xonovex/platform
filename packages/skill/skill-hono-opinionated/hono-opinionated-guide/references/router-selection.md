# router-selection: Pick the Router for the Environment

Pass an explicit `router` to the `Hono` constructor based on deployment (the default `SmartRouter` picks at runtime).

- **RegExpRouter** (`hono/router/reg-exp-router`): fastest throughput (single compiled regex); use for high-throughput persistent servers.

```typescript
import {Hono} from "hono";
import {RegExpRouter} from "hono/router/reg-exp-router";

const server = new Hono({router: new RegExpRouter()}); // high-throughput
```
