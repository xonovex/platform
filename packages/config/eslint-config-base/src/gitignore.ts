import {existsSync} from "node:fs";
import {dirname, join} from "node:path";

export function resolveGitignorePath(startPath: string): string | undefined {
  let directory = dirname(startPath);

  for (;;) {
    const repositoryPath = join(directory, ".git");
    const packagePath = join(directory, "package.json");
    if (existsSync(repositoryPath) && existsSync(packagePath)) {
      const gitignorePath = join(directory, ".gitignore");
      return existsSync(gitignorePath) ? gitignorePath : undefined;
    }

    const parent = dirname(directory);
    if (parent === directory) return undefined;
    directory = parent;
  }
}
