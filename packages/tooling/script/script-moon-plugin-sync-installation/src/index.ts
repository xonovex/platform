#!/usr/bin/env node
import {dirname} from "node:path";
import {fileURLToPath} from "node:url";
import {nodeFileSystem} from "@xonovex/script-moon-common/file-system";
import {logError} from "@xonovex/script-moon-common/logging";
import {findWorkspaceRoot} from "@xonovex/script-moon-common/workspace";
import {runCli} from "./cli.js";

try {
  const root = findWorkspaceRoot(dirname(fileURLToPath(import.meta.url)));
  process.exitCode = runCli(process.argv.slice(2), root, nodeFileSystem);
} catch (error: unknown) {
  logError(error instanceof Error ? error.message : String(error));
  process.exitCode = 1;
}
