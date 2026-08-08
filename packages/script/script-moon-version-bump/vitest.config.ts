import {defineConfig} from "vitest/config";

export default defineConfig({
  test: {
    coverage: {
      // The gaps against these floors are defaultIo in src/file-transaction.ts
      // and defaultDependencies in src/version-bump.ts: the adapters onto a real
      // disk and a real git, which the unit tier supplies fakes for and
      // test/specs/integration drives.
      thresholds: {
        "src/file-transaction.ts": {
          statements: 95,
          branches: 93,
          functions: 71,
          lines: 95,
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
          statements: 86,
          branches: 85,
          functions: 50,
          lines: 87,
        },
      },
    },
  },
});
