import {parseCliArgs} from "@xonovex/script-moon-common/cli-args";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import type {GitRepository} from "@xonovex/script-moon-common/git";
import type {MoonProject} from "@xonovex/script-moon-common/moon-query";
import {findWorkspaceRoot} from "@xonovex/script-moon-common/workspace";
import {detectVersionChanges} from "./detect.js";

export interface VersionDetectDependencies {
  readonly openRepository: (
    rootDir: string,
  ) => Pick<GitRepository, "resolveCommit" | "readFile">;
  readonly listProjects: (rootDir: string) => readonly MoonProject[];
  readonly fs: FileSystem;
}

const cliSpec = {
  name: "moon-version-detect",
  description: "Detect moon projects with version changes since a git ref",
  options: {
    ref: {
      type: "string",
      short: "r",
      description: "Git ref to compare against (default: HEAD~1)",
    },
  },
} as const;

// Prints the ids of the projects whose publishable version differs from the
// version at the ref, as one JSON array.
export const main = (
  argv: readonly string[],
  startDirectory: string,
  dependencies: VersionDetectDependencies,
): number => {
  const args = parseCliArgs(cliSpec, argv);
  if (args.kind === "help") {
    console.log(args.text);
    return 0;
  }
  const ref =
    (args.values.ref as string | undefined) ?? args.positionals[0] ?? "HEAD~1";
  const rootDir = findWorkspaceRoot(startDirectory, undefined, dependencies.fs);
  const repository = dependencies.openRepository(rootDir);
  const changed = detectVersionChanges(
    rootDir,
    repository.resolveCommit(ref),
    dependencies.listProjects(rootDir),
    repository,
    dependencies.fs,
  );
  console.log(JSON.stringify(changed));
  return 0;
};
