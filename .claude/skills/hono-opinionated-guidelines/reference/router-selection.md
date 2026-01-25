# router-selection: Choosing the Right Router for Your Environment

**Guideline:** Select the appropriate Hono router based on deployment environment, optimizing for cold start times in serverless/edge or throughput in traditional servers.

**Rationale:** Hono provides multiple router implementations: RegExpRouter (fastest throughput via single regex), LinearRouter (optimized for serverless cold starts), SmartRouter (default, supports all patterns), and PatternRouter (smallest footprint). Choosing correctly impacts cold starts, request throughput, and memory usage.

**Example:**

```typescript
import {Hono} from "hono";
import {LinearRouter} from "hono/router/linear-router";
import {RegExpRouter} from "hono/router/reg-exp-router";

// High-throughput traditional server
const app = new Hono({
  router: new RegExpRouter(),
});

// Serverless/edge with fast initialization
const serverlessApp = new Hono({
  router: new LinearRouter(),
});

// Default SmartRouter (no configuration needed)
const defaultApp = new Hono();
```

**Techniques:**
- Identify deployment environment: serverless/edge (prioritize cold starts) or traditional server
- Use LinearRouter for serverless/edge environments with repeated initialization
- Use RegExpRouter for high-throughput APIs with persistent connections
- Use SmartRouter (default) when unsure or mixing pattern types
- Consider route complexity: wildcards, optional params, regex patterns
- Profile cold start times and request throughput for your specific use case
