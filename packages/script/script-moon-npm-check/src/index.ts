#!/usr/bin/env node
import {existsSync, readFileSync} from "node:fs";
import {join} from "node:path";
import {parseCliArgs} from "@xonovex/script-moon-common/cli-args";
import {
  logError,
  logInfo,
  logSuccess,
} from "@xonovex/script-moon-common/logging";
import {readPkg} from "@xonovex/script-moon-common/package-json";
import {packagedFilePaths, validatePackedPackage} from "./packed-package.js";
import {validateDeclaredFiles, validatePackage} from "./validate.js";

parseCliArgs({
  name: "moon-npm-check",
  description: "Validate a package is ready for npm publishing",
});

const packageJsonPath = join(process.cwd(), "package.json");
const packageRoot = process.cwd();

let pkg;
try {
  pkg = readPkg(packageJsonPath);
} catch (error: unknown) {
  logError(`Failed to parse package.json: ${String(error)}`);
  process.exit(1);
}

if (pkg.private === true) {
  logInfo(`Skipping private package ${pkg.name ?? "(unnamed)"}`);
  process.exit(0);
}

let packedFiles: readonly string[];
try {
  packedFiles = packagedFilePaths(packageRoot);
} catch (error: unknown) {
  logError(`Failed to inspect packed package: ${String(error)}`);
  process.exit(1);
}

const errors = [
  ...validatePackage(pkg),
  ...validateDeclaredFiles(pkg, (file) => existsSync(join(packageRoot, file))),
  ...validatePackedPackage(pkg, packedFiles, (file) =>
    readFileSync(join(packageRoot, file), "utf8"),
  ),
];

if (errors.length > 0) {
  logError(`\n${pkg.name ?? packageJsonPath} is not ready for publishing:\n`);
  for (const error of errors) {
    logError(`  - ${error}`);
  }
  process.exit(1);
}

logSuccess(
  `${pkg.name ?? "unknown"}@${pkg.version ?? "unknown"} is ready for publishing`,
);
