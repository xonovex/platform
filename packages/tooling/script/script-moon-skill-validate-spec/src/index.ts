#!/usr/bin/env node
import {nodeFileSystem} from "@xonovex/script-moon-common/adapters/file-system-node";
import {processOutput} from "@xonovex/script-moon-common/adapters/output-node";
import {
  DRIFT_LINT_MODE_ENV,
  resolveDriftLintMode,
} from "@xonovex/script-moon-common/drift-budgets";
import {main} from "./validate-skill.js";

try {
  process.exitCode = main(process.argv.slice(2), process.cwd(), {
    fs: nodeFileSystem,
    output: processOutput,
    driftMode: resolveDriftLintMode(process.env[DRIFT_LINT_MODE_ENV]),
  });
} catch (error: unknown) {
  processOutput.stderr(
    `Error: ${error instanceof Error ? error.message : String(error)}`,
  );
  process.exitCode = 1;
}
