#!/usr/bin/env node
import {execFileSync} from "node:child_process";
import {resolveExecutable} from "@xonovex/script-moon-common/adapters/executable-node";
import {nodeFileSystem} from "@xonovex/script-moon-common/adapters/file-system-node";
import {processOutput} from "@xonovex/script-moon-common/adapters/output-node";
import {main} from "./npm-publish.js";
import {isPublished, publishArgs} from "./publish.js";

process.exitCode = main(process.argv.slice(2), process.cwd(), {
  fs: nodeFileSystem,
  output: processOutput,
  isPublished,
  publish: (dryRun) => {
    execFileSync(resolveExecutable("npm"), [...publishArgs(dryRun)], {
      stdio: "inherit",
    });
  },
});
