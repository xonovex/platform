import {mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach, describe, expect, it} from "vitest";
import {resolveGitignorePath} from "../../../src/gitignore.js";

describe("resolveGitignorePath", () => {
  const temporaryDirectories: string[] = [];

  afterEach(() => {
    for (const directory of temporaryDirectories) {
      rmSync(directory, {recursive: true, force: true});
    }
    temporaryDirectories.length = 0;
  });

  function repository(includeGitignore: boolean): {
    root: string;
    sourcePath: string;
  } {
    const root = mkdtempSync(join(tmpdir(), "eslint-config-base-"));
    temporaryDirectories.push(root);
    mkdirSync(join(root, ".git"));
    mkdirSync(join(root, "src"));
    writeFileSync(join(root, "package.json"), "{}");
    if (includeGitignore) writeFileSync(join(root, ".gitignore"), "dist\n");
    return {root, sourcePath: join(root, "src", "index.js")};
  }

  it("finds the repository gitignore from a nested source path", () => {
    const {root, sourcePath} = repository(true);

    const result = resolveGitignorePath(sourcePath);

    expect(result).toBe(join(root, ".gitignore"));
  });

  it("returns undefined when a repository has no gitignore", () => {
    const {sourcePath} = repository(false);

    const result = resolveGitignorePath(sourcePath);

    expect(result).toBeUndefined();
  });

  it("returns undefined outside a Git repository", () => {
    const root = mkdtempSync(join(tmpdir(), "eslint-config-base-"));
    temporaryDirectories.push(root);
    const sourcePath = join(root, "src", "index.js");

    const result = resolveGitignorePath(sourcePath);

    expect(result).toBeUndefined();
  });
});
