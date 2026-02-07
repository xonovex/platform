import baseConfig from "./src/index.ts";
import {defineConfig} from "eslint/config";

export default defineConfig(baseConfig, {
  languageOptions: {
    parserOptions: {
      tsconfigRootDir: import.meta.dirname,
    },
  },
});
