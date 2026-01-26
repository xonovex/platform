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

## License

MIT
