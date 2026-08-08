import {defineConfig} from "vitest/config";

export default defineConfig({
  test: {
    coverage: {
      thresholds: {
        // The scoring rules the trigger evaluators depend on live here and are
        // driven line by line without a process, so this file carries the floor
        // src/trigger-process.ts used to. What remains there is spawn plumbing the
        // unit tier cannot reach; test/specs/integration/trigger-process.test.ts
        // covers it, which is why the project floors sit where they do.
        "src/trigger-scan.ts": {
          statements: 95,
          branches: 90,
          functions: 100,
          lines: 95,
        },
        "src/validation.ts": {
          statements: 95,
          branches: 80,
          functions: 100,
          lines: 100,
        },
      },
    },
  },
});
