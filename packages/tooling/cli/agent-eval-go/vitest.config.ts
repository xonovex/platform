import baseConfig from "@xonovex/vitest-config-node";
import {defineConfig, mergeConfig} from "vitest/config";

export default defineConfig(
  mergeConfig(baseConfig, {
    test: {
      coverage: {
        // bin.ts is the composition root: it holds the spawn, filesystem and
        // process adapters, which only the launched binary exercises.
        exclude: ["src/bin.ts"],
        thresholds: {
          "src/launcher.ts": {
            statements: 95,
            branches: 90,
            functions: 95,
            lines: 95,
          },
        },
      },
    },
  }),
);
