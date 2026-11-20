import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import {infoLine, successLine} from "@xonovex/script-moon-common/logging";
import type {Output} from "@xonovex/script-moon-common/output";
import type {BumpType} from "./bump.js";
import {detectDepUpdates} from "./dep-updates.js";
import {planDependentChanges} from "./dependents-command.js";
import type {FileChange} from "./file-transaction.js";
import type {VersionGit} from "./git.js";
import {
  planLockstep,
  type MemberPlan,
  type WorkspacePackage,
} from "./lockstep.js";
import {planChangelog, serializePackage} from "./package-changes.js";
import {releaseHistory} from "./release.js";

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
  readonly git: VersionGit;
  readonly fs: FileSystem;
  readonly output: Output;
  readonly applyChanges: (changes: readonly FileChange[]) => void;
}

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
    depUpdates: detectDepUpdates(
      options.rootDir,
      member.path,
      member.pkg,
      options.git,
    ),
    changelogFilename: options.changelogPath,
    gitBase: options.gitBase,
    includedTypes: options.includedTypes,
    git: options.git,
    fs: options.fs,
    output: options.output,
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
    isReleased: releaseHistory(options.rootDir, options.git),
  });
  const label = options.dryRun ? "[dry-run] " : "[planned] ";
  const changes: FileChange[] = [];

  for (const member of plan.members) {
    options.output.stdout(
      infoLine(
        `${label}${member.name}: ${member.previousVersion} -> ${plan.targetVersion} (lockstep, base ${plan.baseVersion})`,
      ),
    );
    changes.push(...planMemberChanges(options, member, plan.targetVersion));
  }

  const dependents = options.noDependents ? [] : plan.dependents;
  for (const dependent of dependents) {
    changes.push(...planDependentChanges(options, dependent));
  }

  if (!options.dryRun) options.applyChanges(changes);

  const prefix = options.dryRun ? "[dry-run] Would bump" : "Bumped";
  options.output.stdout(
    successLine(
      `${prefix} ${String(plan.members.length)} package(s) to ${plan.targetVersion}, updated deps in ${String(dependents.length)} file(s).`,
    ),
  );
  return 0;
};

export {runLockstep};
export type {LockstepOptions};
