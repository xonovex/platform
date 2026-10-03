import {join} from "node:path";
import {parseCliArgs} from "@xonovex/script-moon-common/cli-args";
import {
  nodeFileSystem,
  type FileSystem,
} from "@xonovex/script-moon-common/file-system";
import {logInfo, logSuccess} from "@xonovex/script-moon-common/logging";
import {findAllPackageJsonPaths} from "@xonovex/script-moon-common/moon-query";
import {readPkg} from "@xonovex/script-moon-common/package-json";
import {findWorkspaceRoot} from "@xonovex/script-moon-common/workspace";
import {bumpVersion, type BumpType} from "./bump.js";
import {detectDepUpdates} from "./dep-updates.js";
import {planDependentUpdates} from "./dependents-command.js";
import {applyFileChanges, type FileChange} from "./file-transaction.js";
import {getGitVersion, runGit, type GitRunner} from "./git.js";
import {readWorkspacePackages, runLockstep} from "./lockstep-command.js";
import {parseLockstepNames} from "./lockstep.js";
import {planChangelog, serializePackage} from "./package-changes.js";

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

// The two ports that reach outside this process: git reads, and the moon query
// that lists workspace packages. Supplying both lets a caller drive main over a
// plain directory tree with no repository and no moon invocation.
export interface VersionBumpDependencies {
  readonly git: GitRunner;
  readonly listPackagePaths: (rootDir: string) => readonly string[];
  readonly fs: FileSystem;
  // How the planned changes are installed. The transaction owns the ordering and
  // the rollback; supplying it here lets a run be planned without a disk.
  readonly applyChanges: (changes: readonly FileChange[]) => void;
}

export const defaultDependencies: VersionBumpDependencies = {
  git: runGit,
  listPackagePaths: findAllPackageJsonPaths,
  fs: nodeFileSystem,
  applyChanges: (changes) => {
    applyFileChanges(changes);
  },
};

export const main = (
  argv: readonly string[],
  workingDirectory = process.cwd(),
  dependencies: VersionBumpDependencies = defaultDependencies,
): number => {
  const {applyChanges, fs, git, listPackagePaths} = dependencies;
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
        lockstep: {
          type: "string",
          description:
            "Comma-separated packages to move to one shared version in a single write",
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
  const lockstep = values.lockstep as string | undefined;

  if (lockstep !== undefined) {
    validateVersionRequest(exact, bumpType);
    const lockstepRoot = findWorkspaceRoot(workingDirectory, undefined, fs);
    return runLockstep({
      rootDir: lockstepRoot,
      packages: readWorkspacePackages(lockstepRoot, git, listPackagePaths),
      names: parseLockstepNames(lockstep),
      bumpType,
      preid,
      exact,
      dryRun,
      noChangelog,
      noDependents,
      changelogPath,
      gitBase,
      includedTypes,
      git,
    });
  }

  const pkgPath = join(workingDirectory, "package.json");

  if (!fs.isFile(pkgPath)) {
    throw new Error("no package.json found in current directory");
  }

  const pkg = readPkg(pkgPath, fs);
  if (!pkg.name || !pkg.version) {
    throw new Error("package.json is missing a name or version");
  }

  validateVersionRequest(exact, bumpType);

  const oldVersion = pkg.version;
  const rootDir = findWorkspaceRoot(workingDirectory, undefined, fs);
  const gitVersion = getGitVersion(rootDir, pkgPath, git);
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
        fs,
        rootDir,
        packagePaths: listPackagePaths(rootDir),
        packagePath: pkgPath,
        packageName: pkg.name,
        newVersion,
        dryRun,
        noChangelog,
        changelogPath,
        gitBase,
        includedTypes,
        git,
      });
  changes.push(...dependentChanges.changes);

  if (!noChangelog && newVersion !== oldVersion) {
    const changelog = planChangelog({
      rootDir,
      packagePath: pkgPath,
      packageName: pkg.name,
      oldVersion,
      newVersion,
      dryRun,
      depUpdates: detectDepUpdates(rootDir, pkgPath, readPkg(pkgPath, fs)),
      changelogFilename: changelogPath,
      gitBase,
      includedTypes,
      git,
    });
    if (changelog !== undefined) changes.push(changelog);
  }

  if (!dryRun) applyChanges(changes);

  const prefix = dryRun ? "[dry-run] Would bump" : "Bumped";
  logSuccess(
    `${prefix} ${pkg.name} to ${newVersion}, updated deps in ${String(dependentChanges.updated)} file(s).`,
  );
  return 0;
};
