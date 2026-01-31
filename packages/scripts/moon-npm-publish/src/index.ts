#!/usr/bin/env node
import {execSync} from "node:child_process";
import {readFileSync} from "node:fs";
import {
  logInfo,
  parseCliArgs,
  type PackageJson,
} from "@xonovex/moon-scripts-common";

const {values} = parseCliArgs({
  name: "moon-npm-publish",
  description: "Publish a package to npm if not already published",
  options: {
    "dry-run": {
      type: "boolean",
      short: "d",
      description: "Run npm publish in dry-run mode",
    },
  },
});
const dryRun = values["dry-run"] === true;

const {name, version} = JSON.parse(
  readFileSync("package.json", "utf8"),
) as PackageJson & {name: string; version: string};

try {
  execSync(`npm view ${name}@${version} version`, {stdio: "ignore"});
  logInfo(`Skipping ${name}@${version} — already published`);
  process.exit(0);
} catch {
  // Version not found, proceed with publish
}

const publishCmd = dryRun
  ? "npm publish --dry-run --access public"
  : "npm publish --provenance --access public";
// eslint-disable-next-line sonarjs/no-os-command-from-path
execSync(publishCmd, {stdio: "inherit"});
