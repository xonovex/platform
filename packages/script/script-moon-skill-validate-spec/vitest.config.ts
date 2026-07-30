import {defineConfig} from "vitest/config";

export default defineConfig({
  test: {
    include: ["src/**/*.test.ts"],
    coverage: {
      thresholds: {
        // Floors track the suite with the catalog-reading tests skipped; they
        // rise again when those tests run against a migrated catalog.
        "src/validate-skill.ts": {
          statements: 68,
          branches: 45,
          functions: 80,
          lines: 70,
        },
      },
    },
  },
});
