import {builtinModules} from "node:module";
import {GLOB_SRC_TS_WITHOUT_JSX} from "@xonovex/eslint-config-base";
import {defineConfig} from "eslint/config";

export const GLOB_COMPOSITION_ROOT = ["**/src/index.ts"];
export const GLOB_ADAPTERS = ["**/src/adapters/**/*.{ts,mts,cts}"];

// These modules compute and never reach the machine, so a core may use them.
// `parseArgs` is pure only because every caller passes its arguments.
const PURE_BUILTINS = new Set(["path", "path/posix", "path/win32"]);
const PURE_BUILTIN_NAMES: Readonly<Record<string, readonly string[]>> = {
  util: ["parseArgs"],
};

const CORE_MESSAGE =
  "A core module does no I/O: take a port as a parameter and put the Node call in an adapter under src/adapters/, which only src/index.ts imports.";
const ADAPTER_MESSAGE =
  "Only the composition root src/index.ts imports an adapter: take the port as a parameter.";

const bareName = (name: string): string =>
  name.startsWith("node:") ? name.slice("node:".length) : name;

const restrictedBuiltins = [
  ...new Set(builtinModules.map((name) => bareName(name))),
]
  .filter((name) => !PURE_BUILTINS.has(name))
  .flatMap((name) =>
    // Some modules exist only under the node: scheme.
    (builtinModules.includes(name)
      ? [name, `node:${name}`]
      : [`node:${name}`]
    ).map((specifier) => {
      const allowImportNames = PURE_BUILTIN_NAMES[name];
      return allowImportNames === undefined
        ? {name: specifier, message: CORE_MESSAGE}
        : {
            name: specifier,
            allowImportNames: [...allowImportNames],
            message: CORE_MESSAGE,
          };
    }),
  );

const adapterImport = {
  regex: "(^|/)adapters/",
  message: ADAPTER_MESSAGE,
};

/**
 * Enforces ports and adapters in a package whose composition root is
 * `src/index.ts` and whose adapters live in `src/adapters/`: a core module
 * imports no I/O built-in, reads no `process` and imports no adapter, and an
 * adapter imports no other adapter.
 */
export const portsAndAdaptersConfig = defineConfig(
  {
    files: GLOB_SRC_TS_WITHOUT_JSX,
    ignores: [...GLOB_COMPOSITION_ROOT, ...GLOB_ADAPTERS],
    rules: {
      "no-restricted-imports": [
        "error",
        {paths: restrictedBuiltins, patterns: [adapterImport]},
      ],
      "no-restricted-globals": [
        "error",
        {name: "process", message: CORE_MESSAGE},
        {name: "fetch", message: CORE_MESSAGE},
      ],
    },
  },
  {
    files: GLOB_ADAPTERS,
    rules: {
      "no-restricted-imports": [
        "error",
        {
          patterns: [
            adapterImport,
            // A relative import inside src/adapters/ names a sibling adapter.
            {regex: String.raw`^\./`, message: ADAPTER_MESSAGE},
          ],
        },
      ],
    },
  },
);

export default portsAndAdaptersConfig;
