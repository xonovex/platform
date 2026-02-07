# @xonovex/eslint-config-cli

ESLint configuration for CLI tools and scripts.

## Installation

```bash
npm install -D @xonovex/eslint-config-cli
```

## Usage

Create an `eslint.config.js` file:

```javascript
import {cliConfig} from "@xonovex/eslint-config-cli";

export default cliConfig;
```

## Features

Extends `@xonovex/eslint-config-base` with CLI-specific settings:

- Node.js globals enabled
- Relaxed rules for CLI scripts
- Console output allowed

## Export Condition Ordering

The `"import"` condition must appear before `"node"` in the `package.json` exports. This allows script packages (which use the `typescript-script` moon tag without `^:build` dependencies) to lint without this package being built first.

## License

MIT
