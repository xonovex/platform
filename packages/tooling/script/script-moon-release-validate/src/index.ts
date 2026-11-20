#!/usr/bin/env node
import {dirname} from "node:path";
import {fileURLToPath} from "node:url";
import {nodeFileSystem} from "@xonovex/script-moon-common/adapters/file-system-node";
import {nodeGitRepository} from "@xonovex/script-moon-common/adapters/git-node";
import {processOutput} from "@xonovex/script-moon-common/adapters/output-node";
import {main} from "./release-validate.js";

try {
  process.exitCode = main(dirname(fileURLToPath(import.meta.url)), {
    fs: nodeFileSystem,
    output: processOutput,
    openRepository: nodeGitRepository,
  });
} catch (error: unknown) {
  const message = error instanceof Error ? error.message : String(error);
  processOutput.stderr(`Release validation failed unexpectedly: ${message}`);
  process.exitCode = 1;
}
