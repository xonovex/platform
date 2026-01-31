#!/usr/bin/env node
import {execSync} from "node:child_process";
import {existsSync, readFileSync, writeFileSync} from "node:fs";
import {join, relative} from "node:path";
import {
  findAllPackageJsonPaths,
  findWorkspaceRoot,
  logError,
  logInfo,
  logSuccess,
  parseCliArgs,
  readPkg,
  writePkg,
} from "@xonovex/moon-scripts-common";
import {bumpVersion, type BumpType} from "./bump.js";
import {
  determineBumpLevel,
  generateChangelogEntry,
  updateChangelog,
} from "./changelog.js";
import {detectDepUpdates} from "./dep-updates.js";
import {updateDependent} from "./dependents.js";
import {getCommitsSince, getLastVersionRef} from "./git-log.js";
import {getGitVersion} from "./git.js";

const main = (): void => {
  const {values, positionals} = parseCliArgs({
    name: "moon-version-bump",
    description: "Bump package version and update workspace dependents",
    options: {
      type: {
        type: "string",
        short: "t",
        description: "Bump type: patch, minor, or major (default: patch)",
      },
      "dry-run": {
        type: "boolean",
        short: "d",
        description: "Preview changes without writing files",
      },
    },
  });
  const bumpType = ((values.type as string | undefined) ??
    positionals[0] ??
    "patch") as BumpType;
  const dryRun = values["dry-run"] === true;

  const cwd = process.cwd();
  const pkgPath = join(cwd, "package.json");

  if (!existsSync(pkgPath)) {
    logError("No package.json found in current directory.");
    process.exit(1);
  }

  const pkg = readPkg(pkgPath);
  if (!pkg.name || !pkg.version) {
    logError("package.json missing name or version.");
    process.exit(1);
  }

  if (!["patch", "minor", "major"].includes(bumpType)) {
    logError(`Invalid bump type: ${bumpType}. Use patch, minor, or major.`);
    process.exit(1);
  }

  const oldVersion = pkg.version;
  const rootDir = findWorkspaceRoot(cwd);
  const gitVersion = getGitVersion(rootDir, pkgPath);
  let newVersion: string;

  if (gitVersion !== undefined && gitVersion !== oldVersion) {
    logInfo(`${pkg.name}: already bumped (${oldVersion}), skipping.`);
    newVersion = oldVersion;
  } else {
    newVersion = bumpVersion(oldVersion, bumpType);
    if (dryRun) {
      logInfo(`[dry-run] ${pkg.name}: ${oldVersion} -> ${newVersion}`);
    } else {
      pkg.version = newVersion;
      writePkg(pkgPath, pkg);
      logInfo(`${pkg.name}: ${oldVersion} -> ${newVersion}`);
    }
  }

  // Update dependents across workspace
  const allPaths = findAllPackageJsonPaths(rootDir);
  let depsUpdated = 0;

  for (const depPkgPath of allPaths) {
    if (depPkgPath === pkgPath) continue;
    const depPkg = readPkg(depPkgPath);
    const result = updateDependent(
      depPkg,
      depPkgPath,
      pkg.name,
      newVersion,
      () => getGitVersion(rootDir, depPkgPath),
    );
    if (result.depsChanged) {
      if (result.versionBumped) {
        const label = dryRun ? "[dry-run] " : "";
        logInfo(
          `${label}${depPkg.name ?? depPkgPath}: ${String(result.oldVersion)} -> ${String(result.newVersion)} (dependency updated)`,
        );
      }
      if (!dryRun) writePkg(depPkgPath, result.pkg);
      depsUpdated++;
    }
  }

  // Update plugin.json if this is @xonovex/skills
  if (pkg.name === "@xonovex/skills") {
    const pluginJsonPath = join(
      rootDir,
      "packages/plugins/skills/.claude-plugin/plugin.json",
    );
    if (existsSync(pluginJsonPath)) {
      if (dryRun) {
        logInfo(`[dry-run] plugin.json -> ${newVersion}`);
      } else {
        const pluginJson = JSON.parse(
          readFileSync(pluginJsonPath, "utf8"),
        ) as Record<string, unknown>;
        pluginJson.version = newVersion;
        writeFileSync(
          pluginJsonPath,
          JSON.stringify(pluginJson, null, 2) + "\n",
        );
        logInfo(`plugin.json -> ${newVersion}`);
      }
    }
  }

  // Generate changelog entry if version was actually bumped
  if (newVersion !== oldVersion) {
    const pkgDir = relative(rootDir, cwd);
    const sinceRef = getLastVersionRef(rootDir, pkgDir);

    if (sinceRef) {
      let prevVersion: string;
      try {
        const oldPkgJson = execSync(
          `git show ${sinceRef}:${pkgDir}/package.json`,
          {cwd: rootDir, encoding: "utf8"},
        );
        prevVersion = (JSON.parse(oldPkgJson) as {version: string}).version;
      } catch {
        prevVersion = oldVersion;
      }

      const bumpLevel = determineBumpLevel(prevVersion, newVersion);
      const commits = getCommitsSince(rootDir, pkgDir, sinceRef);
      const depUpdates = detectDepUpdates(rootDir, pkgPath);
      const entry = generateChangelogEntry(
        newVersion,
        commits,
        bumpLevel,
        depUpdates,
      );

      if (dryRun) {
        logInfo(`[dry-run] Changelog entry for ${pkg.name}@${newVersion}:`);
        console.log(entry);
      } else {
        const changelogPath = join(cwd, "CHANGELOG.md");
        updateChangelog(changelogPath, pkg.name, entry);
        logInfo(`Updated CHANGELOG.md for ${pkg.name}@${newVersion}`);
      }
    } else {
      logInfo(`${pkg.name}: no previous version found, skipping changelog.`);
    }
  }

  const prefix = dryRun ? "[dry-run] Would bump" : "Bumped";
  logSuccess(
    `${prefix} ${pkg.name} to ${newVersion}, updated deps in ${String(depsUpdated)} file(s).`,
  );
};

main();
