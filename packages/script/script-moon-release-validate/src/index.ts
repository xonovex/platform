#!/usr/bin/env node
import {dirname, resolve} from "node:path";
import {fileURLToPath} from "node:url";
import {validateRelease} from "./release-validation.js";

const scriptDirectory = dirname(fileURLToPath(import.meta.url));
const repositoryRoot = resolve(scriptDirectory, "../../../../..");

try {
  const result = validateRelease(repositoryRoot);
  if (result.failures.length > 0) {
    console.error(
      `Release validation failed: ${String(result.failures.length)}/${String(result.checks)} checks`,
    );
    for (const failure of result.failures) console.error(`- ${failure}`);
    process.exitCode = 1;
  } else {
    console.log(
      `Release validation passed: ${String(result.checks)} checks, ${String(result.pluginPackages)} lockstep packages`,
    );
  }
} catch (error: unknown) {
  const message = error instanceof Error ? error.message : String(error);
  process.stderr.write(`Release validation failed unexpectedly: ${message}\n`);
  process.exitCode = 1;
}
