import {defineConfig} from "vitest/config";

export default defineConfig({
  test: {
    coverage: {
      thresholds: {
        "src/evaluate.ts": {
          statements: 90,
          branches: 70,
          functions: 100,
          lines: 90,
        },
        "src/evaluation-config.ts": {
          statements: 95,
          branches: 90,
          functions: 100,
          lines: 97,
        },
        "src/output-process.ts": {
          statements: 80,
          branches: 68,
          functions: 80,
          lines: 88,
        },
      },
    },
  },
});
