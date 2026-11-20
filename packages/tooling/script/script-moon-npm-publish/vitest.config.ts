import {defineConfig} from "vitest/config";

export default defineConfig({
  test: {
    coverage: {
      // The npm view adapter spawns npm against the registry, which no test
      // tier may reach.
      exclude: ["src/adapters/*-node.ts"],
    },
  },
});
