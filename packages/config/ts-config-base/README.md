# @xonovex/ts-config-base

Base TypeScript configuration for Xonovex projects.

## Installation

```bash
npm install -D @xonovex/ts-config-base
```

## Usage

Extend in your `tsconfig.json`:

```json
{
  "extends": "@xonovex/ts-config-base"
}
```

## Configuration

Provides strict TypeScript settings:

- `strict: true`
- `noUncheckedIndexedAccess: true`
- `exactOptionalPropertyTypes: true`
- ES2022 target
- ESM module resolution

## License

MIT
