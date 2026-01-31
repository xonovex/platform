#!/usr/bin/env node
import {existsSync} from "node:fs";
import {join} from "node:path";
import {
  logError,
  logInfo,
  logSuccess,
  logWarning,
  readPkg,
} from "@xonovex/moon-scripts-common";
import {validatePackage} from "./validate.js";

const packageJsonPath = join(process.cwd(), "package.json");

let pkg;
try {
  pkg = readPkg(packageJsonPath);
} catch (error: unknown) {
  logError(`Failed to parse package.json: ${String(error)}`);
  process.exit(1);
}

if (pkg.private) {
  logInfo(`Skipping private package ${pkg.name ?? "(unnamed)"}`);
  process.exit(0);
}

const errors = validatePackage(pkg);

if (pkg.files && Array.isArray(pkg.files)) {
  for (const file of pkg.files) {
    if (!existsSync(join(process.cwd(), String(file)))) {
      logWarning(
        `  Warning: file "${String(file)}" does not exist yet (may be created during build)`,
      );
    }
  }
}

if (errors.length > 0) {
  logError(
    `\n${pkg.name ?? packageJsonPath} is not ready for publishing:\n`,
  );
  for (const error of errors) {
    logError(`  - ${error}`);
  }
  process.exit(1);
}

logSuccess(`${pkg.name ?? "unknown"}@${pkg.version ?? "unknown"} is ready for publishing`);
