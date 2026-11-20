import {nodeFileSystem} from "@xonovex/script-moon-common/adapters/file-system-node";
import {processOutput} from "@xonovex/script-moon-common/adapters/output-node";
import type {EvaluatorPorts} from "../harness-ports.js";
import {nodeProcessRunner} from "./process-runner-node.js";
import {nodeScratch} from "./scratch-node.js";

/** The ports of an evaluator run on this host. */
export const nodeEvaluatorPorts: EvaluatorPorts = {
  fs: nodeFileSystem,
  scratch: nodeScratch,
  runProcess: nodeProcessRunner,
  environment: process.env,
  output: processOutput,
  platform: process.platform,
};
