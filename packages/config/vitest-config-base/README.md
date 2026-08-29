# @xonovex/vitest-config-base

Use this package to apply the shared Xonovex Vitest configuration to a project.

## Installation

Install the package as a development dependency.

```bash
npm install -D @xonovex/vitest-config-base
```

## Usage

Spread `baseVitestConfig` into the configuration exported from `vitest.config.ts`.

```typescript
import {baseVitestConfig} from "@xonovex/vitest-config-base";
import {defineConfig} from "vitest/config";

export default defineConfig({
  ...baseVitestConfig,
  // Custom overrides
});
```

## Features

The configuration provides these test defaults:

- TypeScript path mapping support via `vite-tsconfig-paths`
- Sensible defaults for test coverage
- Optimized for monorepo setups

## License

MIT
