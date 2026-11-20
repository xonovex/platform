#!/usr/bin/env node
import {nodeFileSystem} from "@xonovex/script-moon-common/adapters/file-system-node";
import {processOutput} from "@xonovex/script-moon-common/adapters/output-node";
import {fetchHttpClient} from "./adapters/http-fetch.js";
import {gitUpstreamRepository} from "./adapters/upstream-git.js";
import {main} from "./audit.js";

try {
  process.exitCode = await main(process.argv.slice(2), process.cwd(), {
    fs: nodeFileSystem,
    output: processOutput,
    http: fetchHttpClient,
    openUpstream: gitUpstreamRepository,
    today: new Date(),
  });
} catch (error: unknown) {
  const message = error instanceof Error ? error.message : String(error);
  processOutput.stderr(`error: ${message}`);
  process.exitCode = 2;
}
