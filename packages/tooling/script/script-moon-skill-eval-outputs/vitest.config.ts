import {defineConfig} from "vitest/config";

export default defineConfig({
  test: {
    coverage: {
      thresholds: {
        // The uncovered function here is the clock in defaultDependencies.
        "src/evaluate.ts": {
          statements: 95,
          branches: 85,
          functions: 90,
          lines: 95,
        },
        // The uncovered function here is defaultLocateExecutable, which looks
        // up the harness CLI on the real PATH. A unit case answers that through
        // the locateExecutable option instead.
        "src/evaluation-config.ts": {
          statements: 95,
          branches: 90,
          functions: 90,
          lines: 96,
        },
        "src/output-parse.ts": {
          statements: 98,
          branches: 93,
          functions: 100,
          lines: 98,
        },
        "src/output-process.ts": {
          statements: 98,
          branches: 90,
          functions: 100,
          lines: 100,
        },
      },
    },
  },
});
