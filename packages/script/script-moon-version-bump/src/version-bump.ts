import {existsSync, readFileSync} from "node:fs";
import {dirname, join, relative} from "node:path";
import {
  findAllPackageJsonPaths,
  findWorkspaceRoot,
  logInfo,
  logSuccess,
  parseCliArgs,
  readPkg,
  type PackageJson,
} from "@xonovex/script-moon-common";
import {bumpVersion, type BumpType} from "./bump.js";
import {
  determineBumpLevel,
  generateChangelogEntry,
  renderUpdatedChangelog,
  type DepUpdate,
} from "./changelog.js";
import {detectDepUpdates} from "./dep-updates.js";
import {updateDependent} from "./dependents.js";
import {applyFileChanges, type FileChange} from "./file-transaction.js";
import {getCommitsSince, getLastVersionRef} from "./git-log.js";
import {getGitVersion} from "./git.js";

const serializePackage = (pkg: PackageJson): string =>
  JSON.stringify(pkg, null, 2) + "\n";

const planChangelog = (
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
): FileChange | undefined => {
  const pkgDir = relative(rootDir, dirname(pkgPath));
  const filename = changelogFilename ?? "CHANGELOG.md";
  const sinceRef = gitBase ?? getLastVersionRef(rootDir, pkgDir, oldVersion);

  if (!sinceRef) {
    logInfo(`${packageName}: no previous version found, skipping changelog.`);
    return undefined;
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
    return undefined;
  }
  const changelogPath = join(dirname(pkgPath), filename);
  const existing = existsSync(changelogPath)
    ? readFileSync(changelogPath, "utf8")
    : undefined;
  return {
    path: changelogPath,
    content: renderUpdatedChangelog(existing, packageName, entry),
  };
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

interface DependentChangeSet {
  readonly updated: number;
  readonly changes: readonly FileChange[];
}

const planDependentUpdates = (
  options: DependentUpdateOptions,
): DependentChangeSet => {
  let updated = 0;
  const changes: FileChange[] = [];
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
      const label = options.dryRun ? "[dry-run] " : "[planned] ";
      logInfo(
        `${label}${dependent.name ?? dependentPath}: ${String(result.oldVersion)} -> ${String(result.newVersion)} (dependency updated)`,
      );
    }
    if (!options.dryRun) {
      changes.push({
        path: dependentPath,
        content: serializePackage(result.pkg),
      });
    }
    updated += 1;

    if (
      !options.noChangelog &&
      result.versionBumped &&
      result.oldVersion &&
      result.newVersion
    ) {
      const changelog = planChangelog(
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
      if (changelog !== undefined) changes.push(changelog);
    }
  }
  return {updated, changes};
};

const validateVersionRequest = (
  exact: string | undefined,
  bumpType: BumpType,
): void => {
  if (exact !== undefined) {
    if (!/^\d+\.\d+\.\d+(?:-\w+\.\d+)?$/.test(exact)) {
      throw new Error(
        `invalid exact version ${exact}; expected X.Y.Z or X.Y.Z-tag.N`,
      );
    }
    return;
  }
  if (!(["patch", "minor", "major"] as const).includes(bumpType)) {
    throw new Error(
      `invalid bump type ${bumpType}; expected patch, minor, or major`,
    );
  }
};

export const main = (
  argv: readonly string[],
  workingDirectory = process.cwd(),
): number => {
  const {values, positionals} = parseCliArgs(
    {
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
    },
    argv,
  );
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

  const pkgPath = join(workingDirectory, "package.json");

  if (!existsSync(pkgPath)) {
    throw new Error("no package.json found in current directory");
  }

  const pkg = readPkg(pkgPath);
  if (!pkg.name || !pkg.version) {
    throw new Error("package.json is missing a name or version");
  }

  validateVersionRequest(exact, bumpType);

  const oldVersion = pkg.version;
  const rootDir = findWorkspaceRoot(workingDirectory);
  const gitVersion = getGitVersion(rootDir, pkgPath);
  let newVersion: string;
  const changes: FileChange[] = [];

  if (gitVersion !== undefined && gitVersion !== oldVersion) {
    logInfo(`${pkg.name}: already bumped (${oldVersion}), skipping.`);
    newVersion = oldVersion;
  } else {
    newVersion = exact ?? bumpVersion(oldVersion, bumpType, preid);
    if (dryRun) {
      logInfo(`[dry-run] ${pkg.name}: ${oldVersion} -> ${newVersion}`);
    } else {
      changes.push({
        path: pkgPath,
        content: serializePackage({...pkg, version: newVersion}),
      });
      logInfo(`[planned] ${pkg.name}: ${oldVersion} -> ${newVersion}`);
    }
  }

  const dependentChanges = noDependents
    ? {updated: 0, changes: []}
    : planDependentUpdates({
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
  changes.push(...dependentChanges.changes);

  if (!noChangelog && newVersion !== oldVersion) {
    const changelog = planChangelog(
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
    if (changelog !== undefined) changes.push(changelog);
  }

  if (!dryRun) applyFileChanges(changes);

  const prefix = dryRun ? "[dry-run] Would bump" : "Bumped";
  logSuccess(
    `${prefix} ${pkg.name} to ${newVersion}, updated deps in ${String(dependentChanges.updated)} file(s).`,
  );
  return 0;
};
