import {existsSync} from "node:fs";
import {dirname, join} from "node:path";

/**
 * Whether a path is present. The walk takes this as a port so it can be driven
 * from a described tree rather than a real one. A config package is the base of
 * config -> shared -> agent, so it cannot reach the workspace filesystem port and
 * declares its own.
 */
export type PathExists = (path: string) => boolean;

export function resolveGitignorePath(
  startPath: string,
  exists: PathExists = existsSync,
): string | undefined {
  let directory = dirname(startPath);

  for (;;) {
    const repositoryPath = join(directory, ".git");
    const packagePath = join(directory, "package.json");
    if (exists(repositoryPath) && exists(packagePath)) {
      const gitignorePath = join(directory, ".gitignore");
      return exists(gitignorePath) ? gitignorePath : undefined;
    }

    const parent = dirname(directory);
    if (parent === directory) return undefined;
    directory = parent;
  }
}
