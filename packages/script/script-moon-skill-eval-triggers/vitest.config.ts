import {defineConfig} from "vitest/config";

export default defineConfig({
  test: {
    coverage: {
      thresholds: {
        // The uncovered functions are defaultDependencies: the spawning trigger
        // checks and the discard that reaches a real disk.
        "src/evaluate.ts": {
          statements: 96,
          branches: 85,
          functions: 80,
          lines: 96,
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
