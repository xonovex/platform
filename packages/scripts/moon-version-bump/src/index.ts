#!/usr/bin/env node
import {existsSync, readFileSync, writeFileSync} from "node:fs";
import {dirname, join, relative} from "node:path";
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
  type DepUpdate,
} from "./changelog.js";
import {detectDepUpdates} from "./dep-updates.js";
import {updateDependent} from "./dependents.js";
import {getCommitsSince, getLastVersionRef} from "./git-log.js";
import {getGitVersion} from "./git.js";

const generateChangelog = (
  rootDir: string,
  pkgPath: string,
  packageName: string,
  oldVersion: string,
  newVersion: string,
  dryRun: boolean,
  depUpdates?: readonly DepUpdate[],
  changelogFilename?: string,
  gitBase?: string,
): void => {
  const pkgDir = relative(rootDir, dirname(pkgPath));
  const filename = changelogFilename ?? "CHANGELOG.md";
  const sinceRef = gitBase ?? getLastVersionRef(rootDir, pkgDir, oldVersion);

  if (!sinceRef) {
    logInfo(`${packageName}: no previous version found, skipping changelog.`);
    return;
  }

  const bumpLevel = determineBumpLevel(oldVersion, newVersion);
  const commits = getCommitsSince(rootDir, pkgDir, sinceRef);
  const deps = depUpdates ?? detectDepUpdates(rootDir, pkgPath);
  const entry = generateChangelogEntry(newVersion, commits, bumpLevel, deps);

  if (dryRun) {
    logInfo(`[dry-run] Changelog entry for ${packageName}@${newVersion}:`);
    console.log(entry);
  } else {
    const changelogPath = join(dirname(pkgPath), filename);
    updateChangelog(changelogPath, packageName, entry);
    logInfo(`Updated ${filename} for ${packageName}@${newVersion}`);
  }
};

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
      "no-changelog": {
        type: "boolean",
        description: "Skip changelog generation",
      },
      "no-dependents": {
        type: "boolean",
        description: "Skip updating dependent packages",
      },
      "changelog-path": {
        type: "string",
        description: "Custom changelog filename (default: CHANGELOG.md)",
      },
      preid: {
        type: "string",
        description: "Prerelease identifier (e.g. beta → 1.2.4-beta.0)",
      },
      exact: {
        type: "string",
        description: "Set exact version instead of bumping",
      },
      "git-base": {
        type: "string",
        description: "Override git ref for changelog commit range",
      },
    },
  });
  const bumpType = ((values.type as string | undefined) ??
    positionals[0] ??
    "patch") as BumpType;
  const dryRun = values["dry-run"] === true;
  const noChangelog = values["no-changelog"] === true;
  const noDependents = values["no-dependents"] === true;
  const changelogPath = values["changelog-path"] as string | undefined;
  const preid = values.preid as string | undefined;
  const exact = values.exact as string | undefined;
  const gitBase = values["git-base"] as string | undefined;

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

  if (exact) {
    if (!/^\d+\.\d+\.\d+(?:-\w+\.\d+)?$/.test(exact)) {
      logError(
        `Invalid exact version: ${exact}. Expected format: X.Y.Z or X.Y.Z-tag.N`,
      );
      process.exit(1);
    }
  } else if (!["patch", "minor", "major"].includes(bumpType)) {
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
    newVersion = exact ?? bumpVersion(oldVersion, bumpType, preid);
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

  if (!noDependents) {
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

        // Generate changelog for dependents that got version-bumped
        if (
          !noChangelog &&
          result.versionBumped &&
          result.oldVersion &&
          result.newVersion
        ) {
          const depName = depPkg.name ?? depPkgPath;
          const depUpdate: DepUpdate = {name: pkg.name, version: newVersion};
          generateChangelog(
            rootDir,
            depPkgPath,
            depName,
            result.oldVersion,
            result.newVersion,
            dryRun,
            [depUpdate],
            changelogPath,
            gitBase,
          );
        }
      }
    }
  }

  // Update .claude-plugin/plugin.json if present
  const pluginJsonPath = join(cwd, ".claude-plugin", "plugin.json");
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

  // Generate changelog for the primary package
  if (!noChangelog && newVersion !== oldVersion) {
    generateChangelog(
      rootDir,
      pkgPath,
      pkg.name,
      oldVersion,
      newVersion,
      dryRun,
      undefined,
      changelogPath,
      gitBase,
    );
  }

  const prefix = dryRun ? "[dry-run] Would bump" : "Bumped";
  logSuccess(
    `${prefix} ${pkg.name} to ${newVersion}, updated deps in ${String(depsUpdated)} file(s).`,
  );
};

main();
