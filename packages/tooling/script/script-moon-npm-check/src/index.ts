#!/usr/bin/env node
import {nodeFileSystem} from "@xonovex/script-moon-common/adapters/file-system-node";
import {logError} from "@xonovex/script-moon-common/logging";
import {npmPackLister} from "./adapters/pack-lister-npm.js";
import {main} from "./npm-check.js";

try {
  process.exitCode = main(process.argv.slice(2), process.cwd(), {
    fs: nodeFileSystem,
    listPackedFiles: npmPackLister,
  });
} catch (error: unknown) {
  logError(error instanceof Error ? error.message : String(error));
  process.exitCode = 1;
}
