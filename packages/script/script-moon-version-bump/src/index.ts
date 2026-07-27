#!/usr/bin/env node
import {logError} from "@xonovex/script-moon-common/logging";
import {main} from "./version-bump.js";

try {
  process.exitCode = main(process.argv.slice(2));
} catch (error: unknown) {
  logError(error instanceof Error ? error.message : String(error));
  process.exitCode = 1;
}
