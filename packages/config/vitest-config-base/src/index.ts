import tsconfigPaths from "vite-tsconfig-paths";
import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    includeSource: ["src/**/*.{ts,tsx,js,jsx}"],
    exclude: [
      "**/node_modules/**",
      "**/dist/**",
      "**/.next/**",
      "**/build/**",
      "**/.astro/**",
    ],
    coverage: {
      provider: "istanbul",
      include: ["src/**"],
      exclude: [
        "src/**/*.test.*",
        "src/**/*.spec.*",
        "**/*.d.ts",
        "**/*.astro",
      ],
      reporter: ["text", "html", "lcov"],
    },
    passWithNoTests: true,
    fakeTimers: {
      toFake: [
        "setTimeout",
        "clearTimeout",
        "setInterval",
        "clearInterval",
        "performance",
        "Date",
      ],
    },
    // Ensure cleanup runs even on test failure
    teardownTimeout: 10_000,
  },
  plugins: [tsconfigPaths({ projects: ["./tsconfig.json"] })],
  resolve: {
    conditions: ["source"],
  },
});
