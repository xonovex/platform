# @xonovex/prettier-config

Use this package to apply the shared Xonovex Prettier configuration to a project.

## Installation

Install the package as a development dependency.

```bash
npm install -D @xonovex/prettier-config
```

## Usage

Set the `prettier` field in `package.json` to the package name.

```json
{
  "prettier": "@xonovex/prettier-config"
}
```

Alternatively, export the configuration from `.prettierrc.js`.

```javascript
export {default} from "@xonovex/prettier-config";
```

## Included Plugins

The configuration enables these plugins:

- `@ianvs/prettier-plugin-sort-imports` - Import sorting
- `prettier-plugin-astro` - Astro file support
- `prettier-plugin-tailwindcss` - Tailwind CSS class sorting

## License

MIT
