import {execFileSync} from "node:child_process";
import {mkdtempSync, rmSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {processExecutableResolver} from "@xonovex/script-moon-common/adapters/executable-node";
import {nodeFileSystem} from "@xonovex/script-moon-common/adapters/file-system-node";
import {nodeGitRepository} from "@xonovex/script-moon-common/adapters/git-node";
import {afterEach} from "vitest";

/** Resolves a command on the PATH of the test process, as the entry point does. */
export const resolveExecutable = processExecutableResolver(nodeFileSystem);

/** Opens a repository through the git adapter the entry point composes. */
export const openGitRepository = nodeGitRepository(resolveExecutable);

const gitExecutable = resolveExecutable("git");

// Git exports GIT_DIR, GIT_INDEX_FILE, and GIT_AUTHOR_* to the hooks it runs,
// and those take precedence over a repository's own config, so running these
// fixtures under a hook would commit with the caller's identity and resolve
// paths against the outer repository. Both config scopes are neutralized for the
// same reason: otherwise the developer's ~/.gitconfig decides whether commits
// are signed and which hooks path git uses.
const gitEnvironment: NodeJS.ProcessEnv = {
  ...Object.fromEntries(
    Object.entries(process.env).filter(([name]) => !name.startsWith("GIT_")),
  ),
  GIT_CONFIG_GLOBAL: "/dev/null",
  GIT_CONFIG_SYSTEM: "/dev/null",
};

/** Runs git in cwd against the isolated environment and returns trimmed stdout. */
export const git = (cwd: string, args: readonly string[]): string =>
  execFileSync(gitExecutable, [...args], {
    cwd,
    encoding: "utf8",
    env: gitEnvironment,
    stdio: ["ignore", "pipe", "pipe"],
  }).trim();

/**
 * Stages everything under cwd, commits it, and returns the new commit hash. The
 * body becomes a second paragraph, which is how a conventional commit carries
 * its footer.
 */
export const commitAll = (
  cwd: string,
  subject: string,
  body?: string,
): string => {
  git(cwd, ["add", "."]);
  git(cwd, [
    "commit",
    "--quiet",
    "-m",
    subject,
    ...(body === undefined ? [] : ["-m", body]),
  ]);
  return git(cwd, ["rev-parse", "HEAD"]);
};

// Creates temporary directories and removes every one it created after each
// test, so no fixture outlives the case that made it.
const trackedDirectories = (): ((prefix: string) => string) => {
  const roots: string[] = [];

  afterEach(() => {
    for (const root of roots) {
      rmSync(root, {recursive: true, force: true});
    }
    roots.length = 0;
  });

  return (prefix) => {
    const root = mkdtempSync(join(tmpdir(), prefix));
    roots.push(root);
    return root;
  };
};

/** Returns a factory for throwaway directories that are not repositories. */
export const temporaryDirectories = (): ((prefix: string) => string) =>
  trackedDirectories();

/**
 * Returns a factory that creates an initialized repository under a temporary
 * directory. The committer identity and the default branch are set explicitly,
 * so a commit needs nothing from the ambient git configuration.
 */
export const gitRepositories = (): ((prefix: string) => string) => {
  const directory = trackedDirectories();

  return (prefix) => {
    const root = directory(prefix);
    git(root, ["-c", "init.defaultBranch=main", "init", "--quiet"]);
    git(root, ["config", "user.name", "Test Author"]);
    git(root, ["config", "user.email", "test@example.com"]);
    return root;
  };
};
