import {defineConfig} from "vitest/config";

export default defineConfig({
  test: {
    coverage: {
      // src/file-system.ts and src/child-process.ts carry no floor here: they are
      // the adapters onto a real disk and a real process, which only
      // test/specs/integration can drive. Everything built on those ports is held
      // near full coverage instead.
      thresholds: {
        "src/file-system-memory.ts": {
          statements: 100,
          branches: 92,
          functions: 100,
          lines: 100,
        },
        "src/fs.ts": {
          statements: 95,
          branches: 90,
          functions: 100,
          lines: 95,
        },
        "src/executable.ts": {
          statements: 100,
          branches: 77,
          functions: 100,
          lines: 100,
        },
        "src/prose-punctuation.ts": {
          statements: 95,
          branches: 90,
          functions: 100,
          lines: 100,
        },
      },
    },
  },
});
