# @xonovex/vitest-config-node

Use this package to run a Node.js application's tests with the shared Xonovex Vitest defaults.

## Installation

Install the package as a development dependency.

```bash
npm install -D @xonovex/vitest-config-node
```

## Usage

Spread `nodeVitestConfig` into the configuration exported from `vitest.config.ts`.

```typescript
import {nodeVitestConfig} from "@xonovex/vitest-config-node";
import {defineConfig} from "vitest/config";

export default defineConfig({
  ...nodeVitestConfig,
  // Custom overrides
});
```

## Features

The configuration extends `@xonovex/vitest-config-base` with these Node.js settings:

- Node.js environment
- Optimized for server-side testing

## License

MIT
