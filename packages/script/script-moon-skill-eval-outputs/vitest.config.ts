import {defineConfig} from "vitest/config";

export default defineConfig({
  test: {
    coverage: {
      // src/output-harness.ts is the one module without a floor here: it is the
      // spawn plumbing behind the HarnessRunner port, which the unit tier cannot
      // reach. test/specs/integration/evaluate.test.ts drives it through a real
      // binary, and the project floors in moon.yml sit below the total it leaves.
      thresholds: {
        // The uncovered functions here are defaultDependencies: the spawning
        // harness runner, the clock, and the discard that reaches a real disk.
        "src/evaluate.ts": {
          statements: 95,
          branches: 85,
          functions: 90,
          lines: 95,
        },
        "src/evaluation-config.ts": {
          statements: 95,
          branches: 90,
          functions: 100,
          lines: 97,
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
