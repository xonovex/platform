/// <reference types="vitest" />

import baseConfig from "@xonovex/vitest-config-node";
import { defineConfig, mergeConfig } from "vitest/config";

export default defineConfig(mergeConfig(baseConfig, {}));
