# @xonovex/ts-config-node

Use this package to compile a TypeScript application for Node.js.

## Installation

Install the package as a development dependency.

```bash
npm install -D @xonovex/ts-config-node
```

## Usage

Extend the package from `tsconfig.json`, then set the application input and output paths.

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

The configuration extends `@xonovex/ts-config-base` with these Node.js settings:

- Node.js types included
- CommonJS interop enabled
- Optimized for server-side applications

## License

MIT
