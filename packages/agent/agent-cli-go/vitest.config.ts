import baseConfig from "@xonovex/vitest-config-node";
import {defineConfig, mergeConfig} from "vitest/config";

export default defineConfig(
  mergeConfig(baseConfig, {
    test: {
      coverage: {
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
