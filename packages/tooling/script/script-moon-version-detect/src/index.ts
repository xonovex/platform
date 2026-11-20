#!/usr/bin/env node
import {dirname} from "node:path";
import {fileURLToPath} from "node:url";
import {nodeFileSystem} from "@xonovex/script-moon-common/adapters/file-system-node";
import {nodeGitRepository} from "@xonovex/script-moon-common/adapters/git-node";
import {queryMoonProjects} from "@xonovex/script-moon-common/adapters/moon-query-node";
import {logError} from "@xonovex/script-moon-common/logging";
import {main} from "./version-detect.js";

try {
  process.exitCode = main(
    process.argv.slice(2),
    dirname(fileURLToPath(import.meta.url)),
    {
      openRepository: nodeGitRepository,
      listProjects: queryMoonProjects,
      fs: nodeFileSystem,
    },
  );
} catch (error: unknown) {
  logError(error instanceof Error ? error.message : String(error));
  process.exitCode = 1;
}
