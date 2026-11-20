import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import {infoLine} from "@xonovex/script-moon-common/logging";
import type {Output} from "@xonovex/script-moon-common/output";
import {detectDepUpdates} from "./dep-updates.js";
import type {FileChange} from "./file-transaction.js";
import type {VersionGit} from "./git.js";
import {planLockstep, type DependentPlan} from "./lockstep.js";
import {planChangelog, serializePackage} from "./package-changes.js";
import {releaseHistory} from "./release.js";
import {readWorkspacePackages} from "./workspace-packages.js";

/** What planning the files of a dependent reads and where it reports. */
interface DependentContext {
  readonly rootDir: string;
  readonly dryRun: boolean;
  readonly noChangelog: boolean;
  readonly changelogPath: string | undefined;
  readonly gitBase: string | undefined;
  readonly includedTypes: ReadonlySet<string> | undefined;
  readonly git: VersionGit;
  readonly fs: FileSystem;
  readonly output: Output;
}

interface DependentUpdateOptions extends DependentContext {
  readonly packagePaths: readonly string[];
  readonly packageName: string;
  readonly newVersion: string;
}

interface DependentChangeSet {
  readonly updated: number;
  readonly changes: readonly FileChange[];
}

const describeVersion = (dependent: DependentPlan): string => {
  if (dependent.newVersion === undefined) return "version unchanged";
  const move = `${String(dependent.previousVersion)} -> ${dependent.newVersion}`;
  return dependent.withReleaseLine ? `${move}, with its release line` : move;
};

/**
 * Reports one planned dependent and returns its manifest and, when its version
 * moves, its changelog. A dry run returns no manifest and prints the changelog.
 */
const planDependentChanges = (
  context: DependentContext,
  dependent: DependentPlan,
): readonly FileChange[] => {
  const label = context.dryRun ? "[dry-run] " : "[planned] ";
  context.output.stdout(
    infoLine(
      `${label}${dependent.name}: ${String(dependent.referenceCount)} reference(s) updated (${describeVersion(dependent)})`,
    ),
  );
  const changes: FileChange[] = [];
  if (!context.dryRun) {
    changes.push({
      path: dependent.path,
      content: serializePackage(dependent.pkg),
    });
  }
  if (
    context.noChangelog ||
    dependent.previousVersion === undefined ||
    dependent.newVersion === undefined
  ) {
    return changes;
  }
  const changelog = planChangelog({
    rootDir: context.rootDir,
    packagePath: dependent.path,
    packageName: dependent.name,
    oldVersion: dependent.previousVersion,
    newVersion: dependent.newVersion,
    dryRun: context.dryRun,
    depUpdates: detectDepUpdates(
      context.rootDir,
      dependent.path,
      dependent.pkg,
      context.git,
    ),
    changelogFilename: context.changelogPath,
    gitBase: context.gitBase,
    includedTypes: context.includedTypes,
    git: context.git,
    fs: context.fs,
    output: context.output,
  });
  if (changelog !== undefined) changes.push(changelog);
  return changes;
};

// A single bump is a lockstep set of one at an exact version, so its dependents
// follow the lockstep rule: references move, released dependents move with
// their release lines, and the moves propagate to a fixed point.
const planDependentUpdates = (
  options: DependentUpdateOptions,
): DependentChangeSet => {
  const plan = planLockstep({
    packages: readWorkspacePackages(
      options.rootDir,
      options.git,
      options.packagePaths,
      options.fs,
    ),
    names: [options.packageName],
    bumpType: "patch",
    preid: undefined,
    exact: options.newVersion,
    isReleased: releaseHistory(options.rootDir, options.git),
  });
  return {
    updated: plan.dependents.length,
    changes: plan.dependents.flatMap((dependent) =>
      planDependentChanges(options, dependent),
    ),
  };
};

export {planDependentChanges, planDependentUpdates};
export type {DependentChangeSet, DependentContext, DependentUpdateOptions};
