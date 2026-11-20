#!/usr/bin/env node
import {nodeFileSystem} from "@xonovex/script-moon-common/adapters/file-system-node";
import {processOutput} from "@xonovex/script-moon-common/adapters/output-node";
import {errorLine} from "@xonovex/script-moon-common/logging";
import {findWorkspaceRoot} from "@xonovex/script-moon-common/workspace";
import {main} from "./cli.js";

try {
  process.exitCode = main(
    process.argv.slice(2),
    findWorkspaceRoot(import.meta.dirname, nodeFileSystem),
    {fs: nodeFileSystem, output: processOutput},
  );
} catch (error: unknown) {
  processOutput.stderr(
    errorLine(error instanceof Error ? error.message : String(error)),
  );
  process.exitCode = 1;
}
