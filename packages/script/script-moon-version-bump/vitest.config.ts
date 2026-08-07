import {defineConfig} from "vitest/config";

export default defineConfig({
  test: {
    coverage: {
      thresholds: {
        "src/file-transaction.ts": {
          statements: 100,
          branches: 90,
          functions: 100,
          lines: 100,
        },
        "src/git-log.ts": {
          statements: 100,
          branches: 95,
          functions: 100,
          lines: 100,
        },
        "src/index.ts": {
          statements: 100,
          branches: 100,
          functions: 100,
          lines: 100,
        },
        "src/lockstep.ts": {
          statements: 100,
          branches: 88,
          functions: 100,
          lines: 100,
        },
        "src/version-bump.ts": {
          statements: 87,
          branches: 85,
          functions: 66,
          lines: 89,
        },
      },
    },
  },
});
