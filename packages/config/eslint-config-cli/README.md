# @xonovex/eslint-config-cli

Use this package to apply the shared Xonovex ESLint rules for command-line tools and scripts.

## Installation

Install the package as a development dependency.

```bash
npm install -D @xonovex/eslint-config-cli
```

## Usage

Export `cliConfig` from an `eslint.config.js` file.

```javascript
import {cliConfig} from "@xonovex/eslint-config-cli";

export default cliConfig;
```

### Ports and adapters

Add `portsAndAdaptersConfig` in a package whose composition root is `src/index.ts` and whose adapters live in `src/adapters/`. A core module then imports no I/O built-in (only `node:path` and `parseArgs` from `node:util`), reads no `process` or `fetch`, and imports no adapter, and an adapter imports no other adapter.

```javascript
import cliConfig from "@xonovex/eslint-config-cli";
import portsAndAdaptersConfig from "@xonovex/eslint-config-cli/ports-and-adapters";
import {defineConfig} from "eslint/config";

export default defineConfig(cliConfig, portsAndAdaptersConfig);
```

## Features

The configuration extends `@xonovex/eslint-config-base` with these command-line settings:

- Node.js globals enabled
- Relaxed rules for CLI scripts
- Console output allowed

## Export Condition Ordering

Keep the `"import"` condition before `"node"` in the `package.json` exports so script packages can lint without building this package first. These packages use the `typescript-script` Moon tag without `^:build` dependencies.

## License

MIT
