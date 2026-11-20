#!/usr/bin/env node
import {nodeFileSystem} from "@xonovex/script-moon-common/adapters/file-system-node";
import {nodeGitRepository} from "@xonovex/script-moon-common/adapters/git-node";
import {queryMoonProjects} from "@xonovex/script-moon-common/adapters/moon-query-node";
import {logError} from "@xonovex/script-moon-common/logging";
import {applyFileChangesOnDisk} from "./adapters/file-transaction-node.js";
import {main} from "./version-bump.js";

try {
  process.exitCode = main(process.argv.slice(2), process.cwd(), {
    openRepository: nodeGitRepository,
    listProjects: queryMoonProjects,
    fs: nodeFileSystem,
    applyChanges: applyFileChangesOnDisk,
  });
} catch (error: unknown) {
  logError(error instanceof Error ? error.message : String(error));
  process.exitCode = 1;
}
