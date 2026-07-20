import {defineConfig} from "vitest/config";

export default defineConfig({
  test: {
    include: ["src/**/*.test.ts"],
    coverage: {
      thresholds: {
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
