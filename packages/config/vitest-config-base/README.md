# @xonovex/vitest-config-base

Base Vitest configuration for Xonovex projects.

## Installation

```bash
npm install -D @xonovex/vitest-config-base
```

## Usage

In your `vitest.config.ts`:

```typescript
import {baseVitestConfig} from "@xonovex/vitest-config-base";
import {defineConfig} from "vitest/config";

export default defineConfig({
  ...baseVitestConfig,
  // Custom overrides
});
```

## Features

- TypeScript path mapping support via `vite-tsconfig-paths`
- Sensible defaults for test coverage
- Optimized for monorepo setups

## License

MIT
