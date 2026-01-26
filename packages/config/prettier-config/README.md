# @xonovex/prettier-config

Shared Prettier configuration for Xonovex projects.

## Installation

```bash
npm install -D @xonovex/prettier-config
```

## Usage

Add to your `package.json`:

```json
{
  "prettier": "@xonovex/prettier-config"
}
```

Or create a `.prettierrc.js` file:

```javascript
export {default} from "@xonovex/prettier-config";
```

## Included Plugins

- `@ianvs/prettier-plugin-sort-imports` - Import sorting
- `prettier-plugin-astro` - Astro file support
- `prettier-plugin-tailwindcss` - Tailwind CSS class sorting

## License

MIT
