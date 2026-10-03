import {defineConfig} from "vitest/config";

export default defineConfig({
  test: {
    coverage: {
      thresholds: {
        "src/validate-skill.ts": {
          statements: 70,
          branches: 49,
          functions: 83,
          lines: 72,
        },
      },
    },
  },
});
