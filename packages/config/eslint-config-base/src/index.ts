import {fileURLToPath} from "node:url";
import jseslint from "@eslint/js";
import eslintViTest from "@vitest/eslint-plugin";
import {configs as eslintRegexpConfigs} from "eslint-plugin-regexp";
import {configs as eslintSecurityConfigs} from "eslint-plugin-security";
import eslintSonarjs from "eslint-plugin-sonarjs";
import eslintUnicorn from "eslint-plugin-unicorn";
import {defineConfig, globalIgnores, includeIgnoreFile} from "eslint/config";
import tseslint from "typescript-eslint";
import {resolveGitignorePath} from "./gitignore.js";

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

export const GLOB_TYPES = ["types/**/*.{ts,tsx,cts,mts}"];

export const GLOB_TEST = [
  "**/*.test-utils.{ts,tsx,cts,mts}",
  "**/*.spec.{ts,tsx,cts,mts}",
  "**/*.test.{ts,tsx,cts,mts}",
  "**/spec.{ts,tsx,cts,mts}",
  "**/test.{ts,tsx,cts,mts}",
];

export const GLOB_SCRIPT = ["scripts/**/*.{ts,cts,mts}"];

export const GLOB_SRC_JS_WITHOUT_JSX = ["**/src/**/*.{js,mjs,cjs}"];
export const GLOB_SRC_TS_WITHOUT_JSX = ["**/src/**/*.{ts,mts,cts}"];

export const GLOB_JS = ["*.{js,jsx,cjs,mjs}", "**/*.{js,jsx,cjs,mjs}"];
export const GLOB_TS = ["*.{ts,tsx,cts,mts}", "**/*.{ts,tsx,cts,mts}"];

const __filename = fileURLToPath(import.meta.url);
const gitignorePath = resolveGitignorePath(__filename);

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

  {
    linterOptions: {
      noInlineConfig: true,
      reportUnusedDisableDirectives: "error",
    },
  },

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
      eslintRegexpConfigs["flat/recommended"],
      eslintUnicorn.configs.recommended,
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
    rules: {
      "@typescript-eslint/no-unused-vars": "off",
      "no-warning-comments": [
        "error",
        {
          terms: [
            "eslint-disable",
            "@ts-ignore",
            "@ts-expect-error",
            "@ts-nocheck",
          ],
          location: "anywhere",
        },
      ],
      "regexp/no-super-linear-backtracking": "off",
      "perfectionist/sort-exports": "off",
      "perfectionist/sort-imports": "off",
      "security/detect-non-literal-regexp": "off",
      "security/detect-non-literal-fs-filename": "off",
      "security/detect-object-injection": "off",
      "security/detect-unsafe-regex": "off",
      "sonarjs/pseudo-random": "off",
      "sonarjs/no-alphabetical-sort": "off",
      "sonarjs/function-return-type": "off",
      "sonarjs/cognitive-complexity": ["error", 30],
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

      // Rules disabled by current project conventions.
      "unicorn/class-reference-in-static-methods": "off",
      "unicorn/consistent-boolean-name": "off",
      "unicorn/consistent-class-member-order": "off",
      "unicorn/consistent-conditional-object-spread": "off",
      "unicorn/logical-assignment-operators": "off",
      "unicorn/max-nested-calls": "off",
      "unicorn/name-replacements": "off",
      "unicorn/no-break-in-nested-loop": "off",
      "unicorn/no-computed-property-existence-check": "off",
      "unicorn/no-declarations-before-early-exit": "off",
      "unicorn/no-duplicate-loops": "off",
      "unicorn/no-global-object-property-assignment": "off",
      "unicorn/no-incorrect-template-string-interpolation": "off",
      "unicorn/no-non-function-verb-prefix": "off",
      "unicorn/no-optional-chaining-on-undeclared-variable": "off",
      "unicorn/no-top-level-assignment-in-function": "off",
      "unicorn/no-top-level-side-effects": "off",
      "unicorn/no-unnecessary-global-this": "off",
      "unicorn/no-unreadable-for-of-expression": "off",
      "unicorn/no-unsafe-string-replacement": "off",
      "unicorn/no-useless-coercion": "off",
      "unicorn/no-useless-concat": "off",
      "unicorn/no-useless-else": "off",
      "unicorn/no-useless-logical-operand": "off",
      "unicorn/no-useless-recursion": "off",
      "unicorn/no-useless-spread": "off",
      "unicorn/operator-assignment": "off",
      "unicorn/prefer-add-event-listener-options": "off",
      "unicorn/prefer-array-from-async": "off",
      "unicorn/prefer-array-from-map": "off",
      "unicorn/prefer-await": "off",
      "unicorn/prefer-boolean-return": "off",
      "unicorn/prefer-continue": "off",
      "unicorn/prefer-direct-iteration": "off",
      "unicorn/prefer-dom-node-replace-children": "off",
      "unicorn/prefer-early-return": "off",
      "unicorn/prefer-else-if": "off",
      "unicorn/prefer-flat-math-min-max": "off",
      "unicorn/prefer-global-number-constants": "off",
      "unicorn/prefer-hoisting-branch-code": "off",
      "unicorn/prefer-iterator-to-array": "off",
      "unicorn/prefer-location-assign": "off",
      "unicorn/prefer-minimal-ternary": "off",
      "unicorn/prefer-number-coercion": "off",
      "unicorn/prefer-number-is-safe-integer": "off",
      "unicorn/prefer-object-define-properties": "off",
      "unicorn/prefer-object-iterable-methods": "off",
      "unicorn/prefer-observer-apis": "off",
      "unicorn/prefer-path2d": "off",
      "unicorn/prefer-promise-with-resolvers": "off",
      "unicorn/prefer-scoped-selector": "off",
      "unicorn/prefer-toggle-attribute": "off",
      "unicorn/prefer-type-literal-last": "off",
      "unicorn/prefer-uint8array-base64": "off",
      "unicorn/prefer-unary-minus": "off",
      "unicorn/prefer-unicode-code-point-escapes": "off",
      "unicorn/prefer-url-href": "off",
      "unicorn/require-array-sort-compare": "off",
      "sonarjs/assertions-in-tests": "off",
      "sonarjs/no-floating-point-equality": "off",
      "sonarjs/no-redundant-optional": "off",
      "sonarjs/no-skipped-tests": "off",
      "sonarjs/no-trivial-assertions": "off",
      "sonarjs/prefer-specific-assertions": "off",
      "sonarjs/super-linear-regex": "off",
    },
  },

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
);
