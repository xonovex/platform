import {existsSync} from "node:fs";
import path from "node:path";
import {fileURLToPath} from "node:url";
import {includeIgnoreFile} from "@eslint/compat";
import jseslint from "@eslint/js";
import eslintViTest from "@vitest/eslint-plugin";
//import eslintFunctional from "eslint-plugin-functional";
//import eslintImportX from "eslint-plugin-import-x";
//import eslintJsdoc from "eslint-plugin-jsdoc";
//import eslintPerfectionist from "eslint-plugin-perfectionist";
//import eslintPrettierRecommended from "eslint-plugin-prettier/recommended";
//import eslintPromise from "eslint-plugin-promise";
import {configs as eslintRegexpConfigs} from "eslint-plugin-regexp";
import {configs as eslintSecurityConfigs} from "eslint-plugin-security";
//import eslintSimpleImportSort from "eslint-plugin-simple-import-sort";
import eslintSonarjs from "eslint-plugin-sonarjs";
import eslintUnicorn from "eslint-plugin-unicorn";
import {defineConfig, globalIgnores} from "eslint/config";
import tseslint from "typescript-eslint";

export const GLOB_CONFIG_TS = [
  ".*.{ts,tsx,cts,mts}",
  "*.setup.{ts,tsx,cts,mts}",
  "*.config.{ts,tsx,cts,mts}",
  "*.options.{ts,tsx,cts,mts}",
  "*.workspace.{ts,tsx,cts,mts}",
  "**/.*.{ts,tsx,cts,mts}",
  "**/*.setup.{ts,tsx,cts,mts}",
  "**/*.config.{ts,tsx,cts,mts}",
  "**/*.options.{ts,tsx,cts,mts}",
];

export const GLOB_CONFIG_JS = [
  ".*.{js,jsx,cjs,mjs}",
  "*.setup.{js,jsx,cjs,mjs}",
  "*.config.{js,jsx,cjs,mjs}",
  "*.options.{js,jsx,cjs,mjs}",
  "*.workspace.{js,jsx,cjs,mjs}",
  "**/.*.{js,jsx,cjs,mjs}",
  "**/*.setup.{js,jsx,cjs,mjs}",
  "**/*.config.{js,jsx,cjs,mjs}",
  "**/*.options.{js,jsx,cjs,mjs}",
  "**/stryker.config.{js,jsx,cjs,mjs}",
  "**/prettier.config.{js,jsx,cjs,mjs}",
];

export const GLOB_FUNCTIONAL_JS = [
  "*.func.{js,mjs,cjs}",
  "**/*.func.{js,mjs,cjs}",
];
export const GLOB_FUNCTIONAL_TS = [
  "*.func.{ts,tsx,cts,mts}",
  "**/*.func.{ts,tsx,cts,mts}",
];

export const GLOB_DECLARATIONS = [
  "*.d.{ts,tsx,cts,mts}",
  "**/*.d.{ts,tsx,cts,mts}",
];

export const GLOB_TYPES = ["types/**/*.{ts,tsx,cts,mts}"];

export const GLOB_TEST = [
  "**/*.test-utils.{ts,tsx,cts,mts}",
  "**/*.spec.{ts,tsx,cts,mts}",
  "**/*.test.{ts,tsx,cts,mts}",
  "**/spec.{ts,tsx,cts,mts}",
  "**/test.{ts,tsx,cts,mts}",
];

export const GLOB_SCRIPT = ["scripts/**/*.{ts,cts,mts}"];

export const GLOB_SRC_JS = ["**/src/**/*.{js,jsx,cjs,mjs}"];
export const GLOB_SRC_TS = ["**/src/**/*.{ts,tsx,cts,mts}"];
export const GLOB_SRC_JS_WITHONLY_JSX = ["**/src/**/*.{jsx}"];
export const GLOB_SRC_TS_WITHONLY_JSX = ["**/src/**/*.{tsx}"];
export const GLOB_SRC_JS_WITHOUT_JSX = ["**/src/**/*.{js,mjs,cjs}"];
export const GLOB_SRC_TS_WITHOUT_JSX = ["**/src/**/*.{ts,mts,cts}"];

export const GLOB_JS = ["*.{js,jsx,cjs,mjs}", "**/*.{js,jsx,cjs,mjs}"];
export const GLOB_TS = ["*.{ts,tsx,cts,mts}", "**/*.{ts,tsx,cts,mts}"];

const __filename = fileURLToPath(import.meta.url);
const __dirname = ((): string | undefined => {
  let dir = path.dirname(__filename);
  while (dir) {
    const hasGit = existsSync(path.join(dir, ".git"));
    const hasPkg = existsSync(path.join(dir, "package.json"));
    if (hasGit && hasPkg) return dir;
    const parent = path.dirname(dir);
    if (parent === dir) return dir;
    dir = parent;
  }
})();

const gitignorePath = __dirname
  ? path.join(__dirname, ".gitignore")
  : undefined;

export const enableTypeCheckedRules = {
  ...tseslint.configs.strictTypeCheckedOnly
    .map((x) => x.rules)

    .reduce((a, b) => ({...a, ...b}), {}),
  ...tseslint.configs.stylisticTypeCheckedOnly
    .map((x) => x.rules)

    .reduce((a, b) => ({...a, ...b}), {}),
  "@typescript-eslint/consistent-type-exports": "error" as const,
  "@typescript-eslint/consistent-type-imports": "error" as const,
};

export const disableTypeCheckedRules = Object.fromEntries(
  Object.keys(enableTypeCheckedRules).map((x) => [x, "off" as const]),
);

export default defineConfig(
  // Global ignores
  globalIgnores([
    // Dependencies and Package Management
    "**/node_modules/**",
    "**/.pnp",
    "**/.pnp.*",
    "**/.yarn/cache/**",
    "**/.yarn/unplugged/**",

    // Build Outputs
    "**/dist/**",
    "**/build/**",
    "**/.output/**",
    "**/out/**",
    "**/.next/**",
    "**/.nuxt/**",
    "**/.nitro/**",

    // Framework-Specific Directories
    "**/.astro/**",
    "**/.svelte-kit/**",
    "**/.tanstack/**",
    "**/.cache/**",
    "**/.docusaurus/**",

    // Mobile Applications
    "**/android/**",
    "**/ios/**",

    // Testing and Coverage
    "**/coverage/**",
    "**/*.coverage",
    "**/coverage*.json",
    "**/coverage*.xml",
    "**/.nyc_output/**",
    "**/test-results/**",
    "**/playwright-report/**",
    "**/blob-report/**",
    "**/playwright/.cache/**",

    // Mutation Testing
    "**/stryker-tmp/**",
    "**/reports/**",

    // Cache and Temporary Files
    "**/.moon/cache/**",
    "**/.parcel-cache/**",
    "**/*.tmp",
    "**/*.bak",
    "**/*.old",
    "**/*.log",
    "**/*storybook.log",

    // Auto-Generated Files
    "**/*.gen.*",
    "**/auto-imports.d.ts",
    "**/routeTree.gen.ts",
    "**/worker-configuration.d.ts",
    "**/*.tsbuildinfo",
    "**/next-env.d.ts",
    "**/api.d.ts",

    // Documentation Build Outputs
    "**/storybook-static/**",
    "**/docs/_build/**",

    // Operating System Files
    "**/.DS_Store",
    "**/*.icloud",

    // Miscellaneous Generated Files
    "**/.assetsignore",
  ]),

  // Git ignores
  gitignorePath ? includeIgnoreFile(gitignorePath) : {},

  // Language options
  {
    languageOptions: {
      ecmaVersion: 2022,
      sourceType: "module",
    },
  },

  // JavaScript
  {
    extends: [jseslint.configs.recommended],
    files: [...GLOB_JS, ...GLOB_SCRIPT],
  },

  // TypeScript
  {
    extends: [
      tseslint.configs.strictTypeChecked,
      tseslint.configs.stylisticTypeChecked,
    ],
    files: [...GLOB_TS],
    rules: {
      ...enableTypeCheckedRules,
      "@typescript-eslint/ban-ts-comment": "off",
    },
  },

  // TypeScript and JavaScript
  {
    extends: [
      // eslintPerfectionist.configs["recommended-natural"],
      eslintRegexpConfigs["flat/recommended"],
      eslintUnicorn.configs.recommended,
      //eslintPromise.configs["flat/recommended"],
      eslintSecurityConfigs.recommended,
      eslintSonarjs.configs.recommended,
    ],
    files: [...GLOB_JS, ...GLOB_TS],
    languageOptions: {
      parser: tseslint.parser,
      parserOptions: {
        projectService: true,
        tsconfigRootDir: import.meta.dirname,
      },
    },
    plugins: {
      // "simple-import-sort": eslintSimpleImportSort,
    },
    rules: {
      "@typescript-eslint/no-unused-vars": "off",
      "regexp/no-super-linear-backtracking": "off",
      "perfectionist/sort-exports": "off",
      "perfectionist/sort-imports": "off",
      // "promise/always-return": "off",
      "security/detect-non-literal-regexp": "off",
      "security/detect-non-literal-fs-filename": "off",
      "security/detect-object-injection": "off",
      "security/detect-unsafe-regex": "off",
      // "simple-import-sort/exports": "warn",
      // "simple-import-sort/imports": "warn",
      "sonarjs/pseudo-random": "off",
      "sonarjs/no-alphabetical-sort": "off",
      "sonarjs/function-return-type": "off",
      "sonarjs/cognitive-complexity": "off",
      "sonarjs/regex-complexity": "off",
      "sonarjs/anchor-precedence": "off",
      "sonarjs/no-commented-code": "off",
      "sonarjs/no-invariant-returns": "off",
      "sonarjs/slow-regex": "off",
      "sonarjs/no-hardcoded-passwords": "off",
      "sonarjs/no-unused-vars": "off",
      "sonarjs/todo-tag": "off",
      "unicorn/prefer-dom-node-append": "off",
      "unicorn/prefer-ternary": "off",
      "unicorn/filename-case": "off",
      "unicorn/import-style": "off",
      "unicorn/no-null": "off",
      "unicorn/no-process-exit": "off",
      "unicorn/prevent-abbreviations": "off",
      "unicorn/no-array-reduce": "off",
      "unicorn/no-array-callback-reference": "off",
      "unicorn/prefer-spread": "off",
      "unicorn/require-module-specifiers": "off",
      "unicorn/number-literal-case": "off",
      "unicorn/template-indent": "off",
    },
  },

  // Typescript type definitions
  {
    files: GLOB_DECLARATIONS,
    rules: {
      "unicorn/no-abusive-eslint-disable": "off",
    },
  },

  // JSDoc TypeScript and JavaScript
  // {
  //   extends: [eslintJsdoc.configs["flat/recommended"]],
  //   files: [...GLOB_JS, ...GLOB_TS],
  //   plugins: {
  //     jsdoc: eslintJsdoc,
  //   },
  //   rules: {
  //     "jsdoc/require-jsdoc": "off",
  //   },
  // },

  // Functional TypeScript and JavaScript
  // {
  //   extends: [eslintFunctional.configs.recommended],
  //   files: [...GLOB_FUNCTIONAL_JS, ...GLOB_FUNCTIONAL_TS],
  // },

  // Packages JavaScript
  // {
  //   extends: [eslintImportX.flatConfigs.recommended],
  //   files: [...GLOB_SRC_JS],
  //   rules: {
  //     "import-x/no-unresolved": "off",
  //   },
  // },

  // Packages TypeScript
  // {
  //   extends: [
  //     eslintImportX.flatConfigs.recommended,
  //     eslintImportX.flatConfigs.typescript,
  //   ],
  //   files: [...GLOB_SRC_TS],
  //   rules: {
  //     "import-x/no-named-as-default-member": "off",
  //     "import-x/no-unresolved": "off",
  //   },
  // },

  // Tests
  {
    extends: [eslintViTest.configs.recommended],
    files: [...GLOB_TEST],
    rules: {
      "@typescript-eslint/unbound-method": "off",
      "@typescript-eslint/no-empty-function": "off",
      "@typescript-eslint/no-unsafe-argument": "off",
      "sonarjs/no-nested-functions": "off",
      "sonarjs/no-hardcoded-secrets": "off",
      "sonarjs/no-identical-functions": "off",
      // Disable false positive for files using custom Playwright fixtures (authTest, etc.)
      "sonarjs/no-empty-test-file": "off",
    },
  },

  // Scripts
  {
    files: GLOB_SCRIPT,
    languageOptions: {
      parser: tseslint.parser,
      parserOptions: {
        project: false,
        projectService: false,
      },
    },
    rules: {
      "no-console": "off",
      "unicorn/no-process-exit": "off",
      "unicorn/prefer-top-level-await": "off",
      ...disableTypeCheckedRules,
    },
  },

  // Configs
  {
    files: [...GLOB_CONFIG_JS, ...GLOB_CONFIG_TS],
    languageOptions: {
      parser: tseslint.parser,
      parserOptions: {
        project: false,
        projectService: false,
      },
    },
    rules: {
      ...disableTypeCheckedRules,
    },
  },

  // Types
  {
    files: GLOB_TYPES,
    languageOptions: {
      parser: tseslint.parser,
      parserOptions: {
        project: false,
        projectService: false,
      },
    },
    rules: {
      ...disableTypeCheckedRules,
    },
  },

  // Prettier
  // eslintPrettierRecommended,
);
