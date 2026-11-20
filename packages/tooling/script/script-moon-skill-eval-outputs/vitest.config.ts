import {defineConfig} from "vitest/config";

export default defineConfig({
  test: {
    coverage: {
      thresholds: {
        "src/evaluate.ts": {
          statements: 100,
          branches: 85,
          functions: 100,
          lines: 100,
        },
        "src/evaluation-config.ts": {
          statements: 100,
          branches: 98,
          functions: 100,
          lines: 100,
        },
        "src/output-parse.ts": {
          statements: 100,
          branches: 95,
          functions: 100,
          lines: 100,
        },
        "src/output-process.ts": {
          statements: 100,
          branches: 94,
          functions: 100,
          lines: 100,
        },
      },
    },
  },
});
