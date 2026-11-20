import {execFileSync} from "node:child_process";
import {mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {dirname, join} from "node:path";
import {afterEach, beforeEach, describe, expect, it} from "vitest";
import {resolveExecutable} from "../../../../src/adapters/executable-node.js";
import {nodeGitRepository} from "../../../../src/adapters/git-node.js";

// nodeGitRepository is the adapter onto a real git, so these cases run it against
// a fresh repository whose history each case controls.
describe("nodeGitRepository", () => {
  let root: string;

  const git = (...args: readonly string[]): string =>
    execFileSync(resolveExecutable("git"), [...args], {
      cwd: root,
      encoding: "utf8",
    });

  const commit = (path: string, text: string, message: string): string => {
    mkdirSync(dirname(join(root, path)), {recursive: true});
    writeFileSync(join(root, path), text);
    git("add", "--all");
    git("commit", "--quiet", "--message", message);
    return git("rev-parse", "HEAD").trim();
  };

  beforeEach(() => {
    root = mkdtempSync(join(tmpdir(), "moon-common-git-"));
    git("init", "--quiet");
    git("config", "user.name", "Test Author");
    git("config", "user.email", "test@example.com");
    git("config", "commit.gpgsign", "false");
  });

  afterEach(() => {
    rmSync(root, {recursive: true, force: true});
  });

  it("resolves a ref and fails for one that names no commit", () => {
    const first = commit("a.txt", "one", "feat: first");
    const repository = nodeGitRepository(root);

    expect(repository.resolveCommit("HEAD")).toBe(first);
    expect(() => repository.resolveCommit("not-a-ref")).toThrow();
  });

  it("reads a file at a commit and reports one the commit lacks", () => {
    const first = commit("pkg/a.txt", "one", "feat: first");
    commit("pkg/a.txt", "two", "fix: second");
    const repository = nodeGitRepository(root);

    expect(repository.readFile(first, "pkg/a.txt")).toBe("one");
    expect(repository.readFile("HEAD", "pkg/a.txt")).toBe("two");
    expect(repository.readFile("HEAD", "pkg/missing.txt")).toBeUndefined();
  });

  it("lists the history of a path and the parent of each commit", () => {
    const first = commit("pkg/a.txt", "one", "feat: first");
    commit("other.txt", "x", "chore: other");
    const third = commit("pkg/a.txt", "two", "fix: second\n\nBody.");
    const repository = nodeGitRepository(root);

    expect(repository.pathHistory("pkg/a.txt")).toEqual([third, first]);
    expect(repository.hasParent(first)).toBe(false);
    expect(repository.hasParent(third)).toBe(true);
    expect(repository.commitsSince(first, "pkg")).toEqual([
      {hash: third, author: "Test Author", message: "fix: second\n\nBody."},
    ]);
    expect(
      repository.commitsSince(undefined, "pkg").map((record) => record.hash),
    ).toEqual([third, first]);
  });

  it("lists the tracked files and the ignored paths", () => {
    writeFileSync(join(root, ".gitignore"), "build/\n");
    mkdirSync(join(root, "build"));
    writeFileSync(join(root, "build", "out.js"), "");
    commit("a.txt", "one", "feat: first");
    const repository = nodeGitRepository(root);

    expect(repository.trackedFiles()).toEqual([".gitignore", "a.txt"]);
    expect(repository.ignoredPaths()).toEqual(["build/"]);
  });
});
