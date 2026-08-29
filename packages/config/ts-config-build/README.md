# @xonovex/ts-config-build

Use this package to compile a TypeScript package for publishing.

## Installation

Install the package as a development dependency.

```bash
npm install -D @xonovex/ts-config-build
```

## Usage

Extend the package from `tsconfig.json`, then set the package input and output paths.

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

The configuration extends `@xonovex/ts-config-base` with these build settings:

- Declaration file generation
- Source maps
- Optimized for package publishing

## License

MIT
