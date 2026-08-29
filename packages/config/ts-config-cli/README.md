# @xonovex/ts-config-cli

Use this package to compile TypeScript command-line tools and scripts.

## Installation

Install the package as a development dependency.

```bash
npm install -D @xonovex/ts-config-cli
```

## Usage

Extend the package from `tsconfig.json`, then set the script input and output paths.

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

The configuration extends `@xonovex/ts-config-base` with these command-line settings:

- Node.js types included
- Optimized for executable scripts

## License

MIT
