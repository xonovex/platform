# @xonovex/vitest-config-node

Vitest configuration for Node.js applications.

## Installation

```bash
npm install -D @xonovex/vitest-config-node
```

## Usage

In your `vitest.config.ts`:

```typescript
import {nodeVitestConfig} from "@xonovex/vitest-config-node";
import {defineConfig} from "vitest/config";

export default defineConfig({
  ...nodeVitestConfig,
  // Custom overrides
});
```

## Features

Extends `@xonovex/vitest-config-base` with Node.js settings:

- Node.js environment
- Optimized for server-side testing

## License

MIT
