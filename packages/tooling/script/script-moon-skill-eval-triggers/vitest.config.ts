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
        // The two functions below the floor here are the defaults behind the
        // executableRuns and resolveExecutablePath ports: one spawns the harness
        // to probe it, the other resolves it on PATH. The unit tier injects both,
        // and test/specs/integration drives the real ones.
        "src/trigger-config.ts": {
          statements: 93,
          branches: 90,
          functions: 75,
          lines: 93,
        },
        "src/trigger-evaluation.ts": {
          statements: 90,
          branches: 85,
          functions: 100,
          lines: 90,
        },
      },
    },
  },
});
