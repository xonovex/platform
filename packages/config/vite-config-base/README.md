# @xonovex/vite-config-base

Base Vite configuration for Xonovex projects.

## Installation

```bash
npm install -D @xonovex/vite-config-base
```

## Usage

In your `vite.config.ts`:

```typescript
import {baseViteConfig} from "@xonovex/vite-config-base";
import {defineConfig} from "vite";

export default defineConfig({
  ...baseViteConfig,
  // Custom overrides
});
```

## Included Plugins

- `@tailwindcss/vite` - Tailwind CSS integration
- `vite-tsconfig-paths` - TypeScript path mapping

## License

MIT
