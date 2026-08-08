import {join} from "node:path";
import {describe, expect, it} from "vitest";
import {resolveGitignorePath, type PathExists} from "../../../src/gitignore.js";

const ROOT = "/repo";
const SOURCE_PATH = join(ROOT, "src", "index.js");

// A tree is just the set of paths the walk asks about.
const present = (...paths: readonly string[]): PathExists => {
  const set = new Set(paths);
  return (path) => set.has(path);
};

// A repository root is the .git and package.json pair, plus the .gitignore the
// walk is looking for.
const repository = (includeGitignore: boolean): PathExists =>
  present(
    join(ROOT, ".git"),
    join(ROOT, "package.json"),
    ...(includeGitignore ? [join(ROOT, ".gitignore")] : []),
  );

describe("resolveGitignorePath", () => {
  it("finds the repository gitignore from a nested source path", () => {
    expect(resolveGitignorePath(SOURCE_PATH, repository(true))).toBe(
      join(ROOT, ".gitignore"),
    );
  });

  it("returns undefined when a repository has no gitignore", () => {
    expect(
      resolveGitignorePath(SOURCE_PATH, repository(false)),
    ).toBeUndefined();
  });

  it("returns undefined outside a Git repository", () => {
    expect(resolveGitignorePath(SOURCE_PATH, present())).toBeUndefined();
  });

  it("walks up from a deeply nested path to the repository root", () => {
    const nested = join(ROOT, "packages", "a", "b", "src", "index.js");

    expect(resolveGitignorePath(nested, repository(true))).toBe(
      join(ROOT, ".gitignore"),
    );
  });

  it("passes over a nested package that carries no repository marker", () => {
    const nested = join(ROOT, "packages", "a", "src", "index.js");
    const exists = present(
      join(ROOT, "packages", "a", "package.json"),
      join(ROOT, ".git"),
      join(ROOT, "package.json"),
      join(ROOT, ".gitignore"),
    );

    expect(resolveGitignorePath(nested, exists)).toBe(join(ROOT, ".gitignore"));
  });
});
