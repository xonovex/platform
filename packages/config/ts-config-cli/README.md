# @xonovex/ts-config-cli

TypeScript configuration for CLI tools and scripts.

## Installation

```bash
npm install -D @xonovex/ts-config-cli
```

## Usage

Extend in your `tsconfig.json`:

```json
{
  "extends": "@xonovex/ts-config-cli",
  "compilerOptions": {
    "outDir": "dist",
    "rootDir": "."
  },
  "include": ["src"]
}
```

## Features

Extends `@xonovex/ts-config-base` with CLI-specific settings:

- Node.js types included
- Optimized for executable scripts

## License

MIT
