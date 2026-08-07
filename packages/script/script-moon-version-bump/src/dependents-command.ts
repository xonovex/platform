import {logInfo} from "@xonovex/script-moon-common/logging";
import {readPkg} from "@xonovex/script-moon-common/package-json";
import {updateDependent} from "./dependents.js";
import type {FileChange} from "./file-transaction.js";
import {getGitVersion, type GitRunner} from "./git.js";
import {planChangelog, serializePackage} from "./package-changes.js";

interface DependentUpdateOptions {
  readonly rootDir: string;
  readonly packagePaths: readonly string[];
  readonly packagePath: string;
  readonly packageName: string;
  readonly newVersion: string;
  readonly dryRun: boolean;
  readonly noChangelog: boolean;
  readonly changelogPath: string | undefined;
  readonly gitBase: string | undefined;
  readonly includedTypes: ReadonlySet<string> | undefined;
  readonly git: GitRunner;
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
  for (const dependentPath of options.packagePaths) {
    if (dependentPath === options.packagePath) continue;
    const dependent = readPkg(dependentPath);
    const result = updateDependent(
      dependent,
      dependentPath,
      options.packageName,
      options.newVersion,
      () => getGitVersion(options.rootDir, dependentPath, options.git),
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
      const changelog = planChangelog({
        rootDir: options.rootDir,
        packagePath: dependentPath,
        packageName: dependent.name ?? dependentPath,
        oldVersion: result.oldVersion,
        newVersion: result.newVersion,
        dryRun: options.dryRun,
        depUpdates: [{name: options.packageName, version: options.newVersion}],
        changelogFilename: options.changelogPath,
        gitBase: options.gitBase,
        includedTypes: options.includedTypes,
        git: options.git,
      });
      if (changelog !== undefined) changes.push(changelog);
    }
  }
  return {updated, changes};
};

export {planDependentUpdates};
export type {DependentChangeSet, DependentUpdateOptions};
