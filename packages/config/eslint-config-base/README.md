# @xonovex/eslint-config-base

Use this package to apply the shared Xonovex ESLint rules to a project.

## Installation

Install the package as a development dependency.

```bash
npm install -D @xonovex/eslint-config-base
```

## Usage

Export `baseConfig` from an `eslint.config.js` file.

```javascript
import {baseConfig} from "@xonovex/eslint-config-base";

export default baseConfig;
```

Spread `baseConfig` before project-specific rules when the project needs overrides.

```javascript
import {baseConfig} from "@xonovex/eslint-config-base";

export default [
  ...baseConfig,
  {
    rules: {
      // Custom rules
    },
  },
];
```

## Included Plugins

The configuration enables these plugins:

- `@typescript-eslint` - TypeScript support
- `eslint-plugin-import` - Import/export linting
- `eslint-plugin-prettier` - Prettier integration
- `eslint-plugin-unicorn` - Various helpful rules
- `eslint-plugin-sonarjs` - Code quality rules
- `eslint-plugin-security` - Security rules
- `eslint-plugin-promise` - Promise best practices
- `eslint-plugin-functional` - Functional programming rules
- `eslint-plugin-perfectionist` - Sorting and ordering
- `eslint-plugin-regexp` - RegExp linting
- `eslint-plugin-jsdoc` - JSDoc linting

## Export Condition Ordering

Keep the `"import"` condition before `"node"` in the `package.json` exports so consumers can use this package without building it first. ESLint uses jiti to load configuration files, and jiti resolves export conditions in object key order. The `"import"` condition resolves to `src/index.ts`. CommonJS consumers fall through to `"node"`, which resolves to `dist/src/index.js`.

## License

MIT
