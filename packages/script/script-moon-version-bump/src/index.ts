#!/usr/bin/env node
import {existsSync} from "node:fs";
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
} from "@xonovex/script-moon-common";
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
  includedTypes?: ReadonlySet<string>,
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
  const entry = generateChangelogEntry(
    newVersion,
    commits,
    bumpLevel,
    deps,
    includedTypes,
  );

  if (dryRun) {
    logInfo(`[dry-run] Changelog entry for ${packageName}@${newVersion}:`);
    console.log(entry);
  } else {
    const changelogPath = join(dirname(pkgPath), filename);
    updateChangelog(changelogPath, packageName, entry);
    logInfo(`Updated ${filename} for ${packageName}@${newVersion}`);
  }
};

interface DependentUpdateOptions {
  readonly rootDir: string;
  readonly packagePath: string;
  readonly packageName: string;
  readonly newVersion: string;
  readonly dryRun: boolean;
  readonly noChangelog: boolean;
  readonly changelogPath: string | undefined;
  readonly gitBase: string | undefined;
  readonly includedTypes: ReadonlySet<string> | undefined;
}

const updateDependents = (options: DependentUpdateOptions): number => {
  let updated = 0;
  for (const dependentPath of findAllPackageJsonPaths(options.rootDir)) {
    if (dependentPath === options.packagePath) continue;
    const dependent = readPkg(dependentPath);
    const result = updateDependent(
      dependent,
      dependentPath,
      options.packageName,
      options.newVersion,
      () => getGitVersion(options.rootDir, dependentPath),
    );
    if (!result.depsChanged) continue;

    if (result.versionBumped) {
      const label = options.dryRun ? "[dry-run] " : "";
      logInfo(
        `${label}${dependent.name ?? dependentPath}: ${String(result.oldVersion)} -> ${String(result.newVersion)} (dependency updated)`,
      );
    }
    if (!options.dryRun) writePkg(dependentPath, result.pkg);
    updated += 1;

    if (
      !options.noChangelog &&
      result.versionBumped &&
      result.oldVersion &&
      result.newVersion
    ) {
      generateChangelog(
        options.rootDir,
        dependentPath,
        dependent.name ?? dependentPath,
        result.oldVersion,
        result.newVersion,
        options.dryRun,
        [{name: options.packageName, version: options.newVersion}],
        options.changelogPath,
        options.gitBase,
        options.includedTypes,
      );
    }
  }
  return updated;
};

const validateVersionRequest = (
  exact: string | undefined,
  bumpType: BumpType,
): void => {
  if (exact !== undefined) {
    if (!/^\d+\.\d+\.\d+(?:-\w+\.\d+)?$/.test(exact)) {
      logError(
        `Invalid exact version: ${exact}. Expected format: X.Y.Z or X.Y.Z-tag.N`,
      );
      process.exit(1);
    }
    return;
  }
  if (!(["patch", "minor", "major"] as const).includes(bumpType)) {
    logError(`Invalid bump type: ${bumpType}. Use patch, minor, or major.`);
    process.exit(1);
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
      "include-types": {
        type: "string",
        description:
          "Comma-separated list of conventional commit types to include (default: feat,fix,refactor,perf,docs)",
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
  const includeTypesRaw = values["include-types"] as string | undefined;
  const includedTypes = includeTypesRaw
    ? new Set(includeTypesRaw.split(",").map((s) => s.trim()))
    : undefined;

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

  validateVersionRequest(exact, bumpType);

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

  const depsUpdated = noDependents
    ? 0
    : updateDependents({
        rootDir,
        packagePath: pkgPath,
        packageName: pkg.name,
        newVersion,
        dryRun,
        noChangelog,
        changelogPath,
        gitBase,
        includedTypes,
      });

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
      includedTypes,
    );
  }

  const prefix = dryRun ? "[dry-run] Would bump" : "Bumped";
  logSuccess(
    `${prefix} ${pkg.name} to ${newVersion}, updated deps in ${String(depsUpdated)} file(s).`,
  );
};

main();
