# @xonovex/ts-config-base

Use this package as the strict TypeScript base for a Xonovex project.

## Installation

Install the package as a development dependency.

```bash
npm install -D @xonovex/ts-config-base
```

## Usage

Extend the package from `tsconfig.json`.

```json
{
  "extends": "@xonovex/ts-config-base"
}
```

## Configuration

The base configuration enables these TypeScript settings:

- `strict: true`
- `noUncheckedIndexedAccess: true`
- `exactOptionalPropertyTypes: true`
- ES2022 target
- ESM module resolution

## License

MIT
