#!/usr/bin/env node
import {dirname} from "node:path";
import {fileURLToPath} from "node:url";
import {processExecutableResolver} from "@xonovex/script-moon-common/adapters/executable-node";
import {nodeFileSystem} from "@xonovex/script-moon-common/adapters/file-system-node";
import {nodeGitRepository} from "@xonovex/script-moon-common/adapters/git-node";
import {queryMoonProjects} from "@xonovex/script-moon-common/adapters/moon-query-node";
import {processOutput} from "@xonovex/script-moon-common/adapters/output-node";
import {errorLine} from "@xonovex/script-moon-common/logging";
import {main} from "./version-detect.js";

const resolveExecutable = processExecutableResolver(nodeFileSystem);

try {
  process.exitCode = main(
    process.argv.slice(2),
    dirname(fileURLToPath(import.meta.url)),
    {
      openRepository: nodeGitRepository(resolveExecutable),
      listProjects: queryMoonProjects(resolveExecutable),
      fs: nodeFileSystem,
      output: processOutput,
    },
  );
} catch (error: unknown) {
  processOutput.stderr(
    errorLine(error instanceof Error ? error.message : String(error)),
  );
  process.exitCode = 1;
}
