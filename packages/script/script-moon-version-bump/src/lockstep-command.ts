import {logInfo, logSuccess} from "@xonovex/script-moon-common/logging";
import {findAllPackageJsonPaths} from "@xonovex/script-moon-common/moon-query";
import {readPkg} from "@xonovex/script-moon-common/package-json";
import type {BumpType} from "./bump.js";
import {detectDepUpdates} from "./dep-updates.js";
import {applyFileChanges, type FileChange} from "./file-transaction.js";
import {getGitVersion} from "./git.js";
import {
  planLockstep,
  type DependentPlan,
  type MemberPlan,
  type WorkspacePackage,
} from "./lockstep.js";
import {planChangelog, serializePackage} from "./package-changes.js";

interface LockstepOptions {
  readonly rootDir: string;
  readonly packages: readonly WorkspacePackage[];
  readonly names: readonly string[];
  readonly bumpType: BumpType;
  readonly preid: string | undefined;
  readonly exact: string | undefined;
  readonly dryRun: boolean;
  readonly noChangelog: boolean;
  readonly noDependents: boolean;
  readonly changelogPath: string | undefined;
  readonly gitBase: string | undefined;
  readonly includedTypes: ReadonlySet<string> | undefined;
}

const readWorkspacePackages = (rootDir: string): readonly WorkspacePackage[] =>
  findAllPackageJsonPaths(rootDir).map((path) => ({
    path,
    pkg: readPkg(path),
    headVersion: getGitVersion(rootDir, path),
  }));

const planMemberChanges = (
  options: LockstepOptions,
  member: MemberPlan,
  targetVersion: string,
): readonly FileChange[] => {
  const changes: FileChange[] = [];
  if (!options.dryRun) {
    changes.push({path: member.path, content: serializePackage(member.pkg)});
  }
  if (options.noChangelog || member.previousVersion === targetVersion) {
    return changes;
  }
  const changelog = planChangelog({
    rootDir: options.rootDir,
    packagePath: member.path,
    packageName: member.name,
    oldVersion: member.baseVersion,
    newVersion: targetVersion,
    dryRun: options.dryRun,
    depUpdates: detectDepUpdates(options.rootDir, member.path, member.pkg),
    changelogFilename: options.changelogPath,
    gitBase: options.gitBase,
    includedTypes: options.includedTypes,
  });
  if (changelog !== undefined) changes.push(changelog);
  return changes;
};

const planDependentChanges = (
  options: LockstepOptions,
  dependent: DependentPlan,
): readonly FileChange[] => {
  const changes: FileChange[] = [];
  if (!options.dryRun) {
    changes.push({
      path: dependent.path,
      content: serializePackage(dependent.pkg),
    });
  }
  if (
    options.noChangelog ||
    dependent.previousVersion === undefined ||
    dependent.newVersion === undefined
  ) {
    return changes;
  }
  const changelog = planChangelog({
    rootDir: options.rootDir,
    packagePath: dependent.path,
    packageName: dependent.name,
    oldVersion: dependent.previousVersion,
    newVersion: dependent.newVersion,
    dryRun: options.dryRun,
    depUpdates: detectDepUpdates(
      options.rootDir,
      dependent.path,
      dependent.pkg,
    ),
    changelogFilename: options.changelogPath,
    gitBase: options.gitBase,
    includedTypes: options.includedTypes,
  });
  if (changelog !== undefined) changes.push(changelog);
  return changes;
};

const runLockstep = (options: LockstepOptions): number => {
  const plan = planLockstep({
    packages: options.packages,
    names: options.names,
    bumpType: options.bumpType,
    preid: options.preid,
    exact: options.exact,
  });
  const label = options.dryRun ? "[dry-run] " : "[planned] ";
  const changes: FileChange[] = [];

  for (const member of plan.members) {
    logInfo(
      `${label}${member.name}: ${member.previousVersion} -> ${plan.targetVersion} (lockstep, base ${plan.baseVersion})`,
    );
    changes.push(...planMemberChanges(options, member, plan.targetVersion));
  }

  const dependents = options.noDependents ? [] : plan.dependents;
  for (const dependent of dependents) {
    const versionNote =
      dependent.newVersion === undefined
        ? "version unchanged"
        : `${String(dependent.previousVersion)} -> ${dependent.newVersion}`;
    logInfo(
      `${label}${dependent.name}: ${String(dependent.referenceCount)} reference(s) -> ${plan.targetVersion} (${versionNote})`,
    );
    changes.push(...planDependentChanges(options, dependent));
  }

  if (!options.dryRun) applyFileChanges(changes);

  const prefix = options.dryRun ? "[dry-run] Would bump" : "Bumped";
  logSuccess(
    `${prefix} ${String(plan.members.length)} package(s) to ${plan.targetVersion}, updated deps in ${String(dependents.length)} file(s).`,
  );
  return 0;
};

export {readWorkspacePackages, runLockstep};
export type {LockstepOptions};
