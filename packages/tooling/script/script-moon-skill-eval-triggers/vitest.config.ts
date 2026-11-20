import {defineConfig} from "vitest/config";

export default defineConfig({
  test: {
    coverage: {
      thresholds: {
        "src/evaluate.ts": {
          statements: 100,
          branches: 88,
          functions: 100,
          lines: 100,
        },
        "src/trigger-config.ts": {
          statements: 98,
          branches: 95,
          functions: 100,
          lines: 98,
        },
        "src/trigger-evaluation.ts": {
          statements: 100,
          branches: 100,
          functions: 100,
          lines: 100,
        },
      },
    },
  },
});
