import {defineConfig} from "vitest/config";

export default defineConfig({
  test: {
    coverage: {
      thresholds: {
        "src/validate-skill.ts": {
          statements: 70,
          branches: 50,
          functions: 90,
          lines: 75,
        },
      },
    },
  },
});
