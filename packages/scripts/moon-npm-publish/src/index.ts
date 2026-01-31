#!/usr/bin/env node
import {execSync} from "node:child_process";
import {readFileSync} from "node:fs";
import {logInfo} from "@xonovex/moon-scripts-common";
import type {PackageJson} from "@xonovex/moon-scripts-common";

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

// eslint-disable-next-line sonarjs/no-os-command-from-path
execSync("npm publish --provenance --access public", {stdio: "inherit"});
