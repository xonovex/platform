import baseConfig from "@xonovex/eslint-config-cli";
import vue from "eslint-plugin-vue";
import {defineConfig} from "eslint/config";
import tseslint from "typescript-eslint";

export default defineConfig(baseConfig, vue.configs["flat/essential"], {
  files: [".vitepress/theme/**/*.vue"],
  languageOptions: {
    parserOptions: {parser: tseslint.parser, extraFileExtensions: [".vue"]},
  },
});
