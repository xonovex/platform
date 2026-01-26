# @xonovex/ts-config-build

TypeScript configuration for building packages.

## Installation

```bash
npm install -D @xonovex/ts-config-build
```

## Usage

Extend in your `tsconfig.json`:

```json
{
  "extends": "@xonovex/ts-config-build",
  "compilerOptions": {
    "outDir": "dist",
    "rootDir": "."
  },
  "include": ["src"]
}
```

## Features

Extends `@xonovex/ts-config-base` with build settings:

- Declaration file generation
- Source maps
- Optimized for package publishing

## License

MIT
