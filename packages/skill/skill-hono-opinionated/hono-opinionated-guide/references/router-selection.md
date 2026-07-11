# router-selection: Pick the Router for the Environment

Pass an explicit `router` to the `Hono` constructor based on deployment; the default is `SmartRouter`.

- **RegExpRouter** (`hono/router/reg-exp-router`) — fastest throughput (single compiled regex); use for high-throughput persistent servers.
- **LinearRouter** (`hono/router/linear-router`) — fastest init, no build step; use for serverless/edge where cold start dominates.
- **SmartRouter** (default) — picks among the others at runtime; supports all patterns. Use when unsure or mixing pattern types.
- **PatternRouter** (`hono/router/pattern-router`) — smallest footprint.

```typescript
import {Hono} from "hono";
import {LinearRouter} from "hono/router/linear-router";
import {RegExpRouter} from "hono/router/reg-exp-router";

const server = new Hono({router: new RegExpRouter()}); // high-throughput
const edge = new Hono({router: new LinearRouter()}); // serverless cold start
```
