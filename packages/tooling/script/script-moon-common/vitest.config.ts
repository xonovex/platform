import {defineConfig} from "vitest/config";

export default defineConfig({
  test: {
    coverage: {
      // The host adapters reach a real disk, a real process, or the process
      // environment, which only test/specs/integration can drive, so the unit
      // floors in moon.yml measure the core and the in-memory adapter alone.
      exclude: ["src/adapters/*-node.ts", "src/adapters/child-process.ts"],
    },
  },
});
