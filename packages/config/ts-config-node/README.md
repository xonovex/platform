# @xonovex/ts-config-node

TypeScript configuration for Node.js applications.

## Installation

```bash
npm install -D @xonovex/ts-config-node
```

## Usage

Extend in your `tsconfig.json`:

```json
{
  "extends": "@xonovex/ts-config-node",
  "compilerOptions": {
    "outDir": "dist",
    "rootDir": "."
  },
  "include": ["src"]
}
```

## Features

Extends `@xonovex/ts-config-base` with Node.js settings:

- Node.js types included
- CommonJS interop enabled
- Optimized for server-side applications

## License

MIT
