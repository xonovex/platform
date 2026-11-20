import {defineConfig} from "vitest/config";

export default defineConfig({
  test: {
    coverage: {
      // The adapters reach a real disk, and test/specs/integration drives them.
      exclude: ["src/adapters/*-node.ts"],
    },
  },
});
