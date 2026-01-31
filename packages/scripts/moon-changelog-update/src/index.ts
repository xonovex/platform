#!/usr/bin/env node
import {execSync} from "node:child_process";
import {readFileSync} from "node:fs";
import {join, relative} from "node:path";
import {
  findWorkspaceRoot,
  logError,
  logInfo,
  logSuccess,
  parseCliArgs,
  type PackageJson,
} from "@xonovex/moon-scripts-common";
import {
  determineBumpLevel,
  generateChangelogEntry,
  updateChangelog,
} from "./changelog.js";
import {detectDepUpdates} from "./dep-updates.js";
import {getCommitsSince, getLastVersionRef} from "./git-log.js";

const {values} = parseCliArgs({
  name: "moon-changelog-update",
  description: "Generate CHANGELOG.md entries from git history",
  options: {
    "dry-run": {
      type: "boolean",
      short: "d",
      description: "Print generated entry without writing",
    },
  },
});
const dryRun = values["dry-run"] === true;

const cwd = process.cwd();
const rootDir = findWorkspaceRoot(cwd);
const pkgPath = join(cwd, "package.json");
const pkg = JSON.parse(readFileSync(pkgPath, "utf8")) as PackageJson & {
  name: string;
  version: string;
};

if (!pkg.name || !pkg.version) {
  logError("package.json missing name or version.");
  process.exit(1);
}

const pkgDir = relative(rootDir, cwd);
const sinceRef = getLastVersionRef(rootDir, pkgDir);

if (!sinceRef) {
  logInfo(`${pkg.name}: no previous version found, skipping changelog.`);
  process.exit(0);
}

// Read old version from git at sinceRef
let oldVersion: string;
try {
  const oldPkgJson = execSync(`git show ${sinceRef}:${pkgDir}/package.json`, {
    cwd: rootDir,
    encoding: "utf8",
  });
  oldVersion = (JSON.parse(oldPkgJson) as {version: string}).version;
} catch {
  oldVersion = pkg.version;
}

const bumpLevel = determineBumpLevel(oldVersion, pkg.version);
const commits = getCommitsSince(rootDir, pkgDir, sinceRef);
const depUpdates = detectDepUpdates(rootDir, pkgPath);

const entry = generateChangelogEntry(
  pkg.version,
  commits,
  bumpLevel,
  depUpdates,
);

if (dryRun) {
  logInfo(`[dry-run] Changelog entry for ${pkg.name}@${pkg.version}:`);
  console.log(entry);
} else {
  const changelogPath = join(cwd, "CHANGELOG.md");
  updateChangelog(changelogPath, pkg.name, entry);
  logSuccess(`Updated CHANGELOG.md for ${pkg.name}@${pkg.version}`);
}
