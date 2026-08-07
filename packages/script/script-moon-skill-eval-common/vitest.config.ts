import {defineConfig} from "vitest/config";

export default defineConfig({
  test: {
    coverage: {
      thresholds: {
        "src/trigger-process.ts": {
          statements: 80,
          branches: 65,
          functions: 80,
          lines: 80,
        },
      },
    },
  },
});
