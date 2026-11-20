import {defineConfig} from "vitest/config";

export default defineConfig({
  test: {
    coverage: {
      // The adapters reach the network and run git, which only
      // test/specs/integration drives, so the unit floors in moon.yml measure
      // the core alone.
      exclude: ["src/adapters/*"],
    },
  },
});
