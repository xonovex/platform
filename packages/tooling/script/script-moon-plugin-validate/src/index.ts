#!/usr/bin/env node
import {nodeFileSystem} from "@xonovex/script-moon-common/adapters/file-system-node";
import {processOutput} from "@xonovex/script-moon-common/adapters/output-node";
import {main} from "./cli.js";

process.exitCode = main(process.argv.slice(2), process.cwd(), {
  fs: nodeFileSystem,
  output: processOutput,
});
