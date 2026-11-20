import {join} from "node:path";
import {parseCliArgs} from "@xonovex/script-moon-common/cli-args";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import {infoLine, successLine} from "@xonovex/script-moon-common/logging";
import {
  packageJsonPaths,
  type MoonProject,
} from "@xonovex/script-moon-common/moon-query";
import type {Output} from "@xonovex/script-moon-common/output";
import {readPkg} from "@xonovex/script-moon-common/package-json";
import {findWorkspaceRoot} from "@xonovex/script-moon-common/workspace";
import {bumpVersion, type BumpType} from "./bump.js";
import {detectDepUpdates} from "./dep-updates.js";
import {planDependentUpdates} from "./dependents-command.js";
import type {FileChange} from "./file-transaction.js";
import {getGitVersion, type VersionGit} from "./git.js";
import {runLockstep} from "./lockstep-command.js";
import {parseLockstepNames} from "./lockstep.js";
import {planChangelog, serializePackage} from "./package-changes.js";
import {readWorkspacePackages} from "./workspace-packages.js";

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

// The ports that reach outside this process. Supplying them lets a caller drive
// main over a plain directory tree with no repository and no moon invocation.
export interface VersionBumpDependencies {
  readonly openRepository: (rootDir: string) => VersionGit;
  readonly listProjects: (rootDir: string) => readonly MoonProject[];
  readonly fs: FileSystem;
  readonly output: Output;
  // How the planned changes are installed. The transaction owns the ordering and
  // the rollback; supplying it here lets a run be planned without a disk.
  readonly applyChanges: (changes: readonly FileChange[]) => void;
}

export const main = (
  argv: readonly string[],
  workingDirectory: string,
  dependencies: VersionBumpDependencies,
): number => {
  const {applyChanges, fs, listProjects, openRepository, output} = dependencies;
  const listPackagePaths = (rootDir: string): readonly string[] =>
    packageJsonPaths(rootDir, listProjects(rootDir), fs);
  const args = parseCliArgs(
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
  if (args.kind === "help") {
    output.stdout(args.text);
    return 0;
  }
  const {values, positionals} = args;
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
    const lockstepRoot = findWorkspaceRoot(workingDirectory, fs);
    const git = openRepository(lockstepRoot);
    return runLockstep({
      rootDir: lockstepRoot,
      packages: readWorkspacePackages(
        lockstepRoot,
        git,
        listPackagePaths(lockstepRoot),
        fs,
      ),
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
      fs,
      output,
      applyChanges,
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
  const rootDir = findWorkspaceRoot(workingDirectory, fs);
  const git = openRepository(rootDir);
  const gitVersion = getGitVersion(rootDir, pkgPath, git);
  let newVersion: string;
  const changes: FileChange[] = [];

  if (gitVersion === undefined || gitVersion === oldVersion) {
    newVersion = exact ?? bumpVersion(oldVersion, bumpType, preid);
    if (dryRun) {
      output.stdout(
        infoLine(`[dry-run] ${pkg.name}: ${oldVersion} -> ${newVersion}`),
      );
    } else {
      changes.push({
        path: pkgPath,
        content: serializePackage({...pkg, version: newVersion}),
      });
      output.stdout(
        infoLine(`[planned] ${pkg.name}: ${oldVersion} -> ${newVersion}`),
      );
    }
  } else {
    output.stdout(
      infoLine(`${pkg.name}: already bumped (${oldVersion}), skipping.`),
    );
    newVersion = oldVersion;
  }

  const dependentChanges = noDependents
    ? {updated: 0, changes: []}
    : planDependentUpdates({
        fs,
        rootDir,
        packagePaths: listPackagePaths(rootDir),
        packageName: pkg.name,
        newVersion,
        dryRun,
        noChangelog,
        changelogPath,
        gitBase,
        includedTypes,
        git,
        output,
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
      depUpdates: detectDepUpdates(rootDir, pkgPath, readPkg(pkgPath, fs), git),
      changelogFilename: changelogPath,
      gitBase,
      includedTypes,
      git,
      fs,
      output,
    });
    if (changelog !== undefined) changes.push(changelog);
  }

  if (!dryRun) applyChanges(changes);

  const prefix = dryRun ? "[dry-run] Would bump" : "Bumped";
  output.stdout(
    successLine(
      `${prefix} ${pkg.name} to ${newVersion}, updated deps in ${String(dependentChanges.updated)} file(s).`,
    ),
  );
  return 0;
};
