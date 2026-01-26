# @xonovex/eslint-config-base

Base ESLint configuration for Xonovex projects.

## Installation

```bash
npm install -D @xonovex/eslint-config-base
```

## Usage

Create an `eslint.config.js` file:

```javascript
import {baseConfig} from "@xonovex/eslint-config-base";

export default baseConfig;
```

Or extend with custom rules:

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

## License

MIT
