import {defineConfig} from "vitest/config";

export default defineConfig({
  test: {
    include: ["src/**/*.test.ts"],
    coverage: {
      thresholds: {
        "src/audit.ts": {
          statements: 55,
          branches: 40,
          functions: 50,
          lines: 60,
        },
      },
    },
  },
});
