import {defineConfig} from "vitest/config";

export default defineConfig({
  test: {
    coverage: {
      thresholds: {
        "src/evaluate.ts": {
          statements: 90,
          branches: 55,
          functions: 100,
          lines: 90,
        },
        "src/trigger-config.ts": {
          statements: 95,
          branches: 90,
          functions: 100,
          lines: 95,
        },
        "src/trigger-evaluation.ts": {
          statements: 90,
          branches: 85,
          functions: 100,
          lines: 90,
        },
      },
    },
  },
});
