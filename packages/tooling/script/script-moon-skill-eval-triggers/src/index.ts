#!/usr/bin/env node
import {nodeEvaluatorPorts} from "@xonovex/script-moon-skill-eval-common/adapters/evaluator-ports-node";
import {runEvaluator} from "@xonovex/script-moon-skill-eval-common/report";
import {main} from "./evaluate.js";

process.exitCode = await runEvaluator(
  () => main(process.argv.slice(2), process.cwd(), nodeEvaluatorPorts),
  nodeEvaluatorPorts.output,
);
