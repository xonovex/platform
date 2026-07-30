import {defineConfig} from "vitest/config";

export default defineConfig({
  test: {
    include: ["src/**/*.test.ts"],
    coverage: {
      thresholds: {
        "src/file-transaction.ts": {
          statements: 100,
          branches: 90,
          functions: 100,
          lines: 100,
        },
        "src/git-log.ts": {
          statements: 95,
          branches: 85,
          functions: 100,
          lines: 100,
        },
        "src/version-bump.ts": {
          statements: 70,
          branches: 60,
          functions: 55,
          lines: 70,
        },
      },
    },
  },
});
