import {fileURLToPath} from "node:url";

const sortImportsPath = fileURLToPath(
  import.meta.resolve("@ianvs/prettier-plugin-sort-imports"),
);
const astroPath = fileURLToPath(import.meta.resolve("prettier-plugin-astro"));
const tailwindPath = fileURLToPath(
  import.meta.resolve("prettier-plugin-tailwindcss"),
);

/** @type {import("prettier").Config} */
export default {
  // Core formatting options
  semi: true,
  tabWidth: 2,
  useTabs: false,
  singleQuote: false,
  bracketSameLine: true,
  bracketSpacing: false,

  // Plugin configuration
  // Note: The Tailwind CSS plugin must be loaded last
  plugins: [sortImportsPath, astroPath, tailwindPath],

  // Import order configuration
  importOrderParserPlugins: ["typescript", "jsx", "decorators-legacy"],
  importOrderTypeScriptVersion: "5.0.0",
};
