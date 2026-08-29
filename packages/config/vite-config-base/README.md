# @xonovex/vite-config-base

Use this package to apply the shared Xonovex Vite configuration to a project.

## Installation

Install the package as a development dependency.

```bash
npm install -D @xonovex/vite-config-base
```

## Usage

Spread `baseViteConfig` into the configuration exported from `vite.config.ts`.

```typescript
import {baseViteConfig} from "@xonovex/vite-config-base";
import {defineConfig} from "vite";

export default defineConfig({
  ...baseViteConfig,
  // Custom overrides
});
```

## Included Plugins

The configuration enables these plugins:

- `@tailwindcss/vite` - Tailwind CSS integration
- `vite-tsconfig-paths` - TypeScript path mapping

## License

MIT
