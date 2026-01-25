import baseConfig, {
  GLOB_SRC_JS_WITHOUT_JSX,
  GLOB_SRC_TS_WITHOUT_JSX,
} from "@xonovex/eslint-config-base";
import { defineConfig } from "eslint/config";
import globals from "globals";

export default defineConfig(
  baseConfig,
  // CLI scripts without JSX
  {
    files: [...GLOB_SRC_JS_WITHOUT_JSX, ...GLOB_SRC_TS_WITHOUT_JSX],
    languageOptions: {
      globals: {
        ...globals.node,
      },
    },
    rules: {
      // CLI tools are expected to execute shell commands
      "sonarjs/os-command": "off",
      "security/detect-child-process": "off",
    },
  },
);
