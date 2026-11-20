import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import type {GitRepository} from "@xonovex/script-moon-common/git";
import {findWorkspaceRoot} from "@xonovex/script-moon-common/workspace";
import {readReleaseSnapshot} from "./release-reader.js";
import {validateRelease} from "./release-validation.js";

export interface ReleaseValidateDependencies {
  readonly fs: FileSystem;
  readonly openRepository: (
    rootDir: string,
  ) => Pick<GitRepository, "trackedFiles" | "ignoredPaths">;
}

/** Validates the release inputs of the workspace above startDirectory. */
export const main = (
  startDirectory: string,
  dependencies: ReleaseValidateDependencies,
): number => {
  const repositoryRoot = findWorkspaceRoot(
    startDirectory,
    undefined,
    dependencies.fs,
  );
  const result = validateRelease(
    readReleaseSnapshot(
      repositoryRoot,
      dependencies.fs,
      dependencies.openRepository(repositoryRoot),
    ),
  );
  if (result.failures.length > 0) {
    console.error(
      `Release validation failed: ${String(result.failures.length)}/${String(result.checks)} checks`,
    );
    for (const failure of result.failures) console.error(`- ${failure}`);
    return 1;
  }
  console.log(
    `Release validation passed: ${String(result.checks)} checks, ${String(result.pluginPackages)} lockstep packages`,
  );
  return 0;
};
