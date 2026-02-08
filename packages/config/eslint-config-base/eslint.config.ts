import {defineConfig} from "eslint/config";
import baseConfig from "./src/index.ts";

export default defineConfig(baseConfig, {
  languageOptions: {
    parserOptions: {
      tsconfigRootDir: import.meta.dirname,
    },
  },
});
