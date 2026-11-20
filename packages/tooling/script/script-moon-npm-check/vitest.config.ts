import {defineConfig} from "vitest/config";

export default defineConfig({
  test: {
    coverage: {
      // The npm adapter spawns npm, which only test/specs/integration drives.
      exclude: ["src/adapters/*-npm.ts"],
    },
  },
});
