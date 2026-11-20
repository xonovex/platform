#!/usr/bin/env node
import {execFileSync} from "node:child_process";
import {processExecutableResolver} from "@xonovex/script-moon-common/adapters/executable-node";
import {nodeFileSystem} from "@xonovex/script-moon-common/adapters/file-system-node";
import {processOutput} from "@xonovex/script-moon-common/adapters/output-node";
import {runNpmView} from "./adapters/npm-view-node.js";
import {main} from "./npm-publish.js";
import {isPublished, publishArgs} from "./publish.js";

const resolveExecutable = processExecutableResolver(nodeFileSystem);

process.exitCode = main(process.argv.slice(2), process.cwd(), {
  fs: nodeFileSystem,
  output: processOutput,
  isPublished: (identity) =>
    isPublished(identity, runNpmView(resolveExecutable)),
  publish: (dryRun) => {
    execFileSync(resolveExecutable("npm"), [...publishArgs(dryRun)], {
      stdio: "inherit",
    });
  },
});
