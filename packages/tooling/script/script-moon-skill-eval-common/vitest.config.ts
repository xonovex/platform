import {defineConfig} from "vitest/config";

export default defineConfig({
  test: {
    coverage: {
      // The node adapters spawn a real process or reach a real disk, which only
      // test/specs/integration drives.
      exclude: ["src/adapters/*-node.ts"],
    },
  },
});
