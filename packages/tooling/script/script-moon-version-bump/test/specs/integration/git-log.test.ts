import {mkdirSync, writeFileSync} from "node:fs";
import {join} from "node:path";
import {describe, expect, it} from "vitest";
import {
  findChangelogRange,
  getCommitsSince,
  isIncludedType,
  parseConventionalCommit,
} from "../../../src/git-log.js";
import {commitAll, gitRepositories} from "../../util/git-repository.js";

const repository = gitRepositories();

const createRepository = (): string => {
  const directory = repository("version-git-log-");
  mkdirSync(join(directory, "packages", "example"), {recursive: true});
  return directory;
};

const writePackage = (directory: string, version: string): void => {
  writeFileSync(
    join(directory, "packages", "example", "package.json"),
    `${JSON.stringify({name: "example", version}, null, 2)}\n`,
  );
};

describe("parseConventionalCommit", () => {
  it("should parse type and description", () => {
    expect(parseConventionalCommit("feat: add new feature")).toEqual({
      type: "feat",
      description: "add new feature",
    });
  });

  it("should parse scoped commits", () => {
    expect(parseConventionalCommit("fix(core): resolve bug")).toEqual({
      type: "fix",
      description: "resolve bug",
    });
  });

  it("should return undefined for non-conventional messages", () => {
    expect(parseConventionalCommit("random commit message")).toBeUndefined();
  });

  it("should handle colons in description", () => {
    expect(parseConventionalCommit("feat: support key: value pairs")).toEqual({
      type: "feat",
      description: "support key: value pairs",
    });
  });
});

describe("isIncludedType", () => {
  it("should include feat, fix, refactor, perf, docs", () => {
    for (const t of ["feat", "fix", "refactor", "perf", "docs"]) {
      expect(isIncludedType(t)).toBe(true);
    }
  });

  it("should exclude chore, ci, build, test, style", () => {
    for (const t of ["chore", "ci", "build", "test", "style"]) {
      expect(isIncludedType(t)).toBe(false);
    }
  });

  it("uses an explicit included type set", () => {
    expect(isIncludedType("chore", new Set(["chore"]))).toBe(true);
    expect(isIncludedType("feat", new Set(["chore"]))).toBe(false);
  });
});

describe("git history", () => {
  it("finds the commit containing the preceding package version", () => {
    const directory = createRepository();
    writePackage(directory, "1.0.0");
    const previousVersionCommit = commitAll(directory, "feat: initial package");
    writePackage(directory, "2.0.0");
    commitAll(directory, "chore: release 2.0.0");

    expect(findChangelogRange(directory, "packages/example", "2.0.0")).toEqual({
      since: previousVersionCommit,
    });
  });

  it("falls back to before the package introduction when its version never changed", () => {
    const directory = createRepository();
    writeFileSync(join(directory, "unrelated.txt"), "ok\n");
    commitAll(directory, "chore: seed the repository");
    writePackage(directory, "1.0.0");
    const introduction = commitAll(directory, "feat: initial package");

    expect(findChangelogRange(directory, "packages/example", "1.0.0")).toEqual({
      since: `${introduction}~1`,
    });
    expect(
      findChangelogRange(directory, "packages/missing", "1.0.0"),
    ).toBeUndefined();
  });

  it("covers the whole history when the root commit introduced the package", () => {
    const directory = createRepository();
    writePackage(directory, "1.0.0");
    const introduction = commitAll(directory, "feat: initial package");

    expect(findChangelogRange(directory, "packages/example", "1.0.0")).toEqual({
      since: undefined,
    });
    expect(
      getCommitsSince(directory, "packages/example", undefined),
    ).toMatchObject([{hash: introduction, author: "Test Author"}]);
  });

  it("extracts conventional subject and body messages from package commits", () => {
    const directory = createRepository();
    writePackage(directory, "1.0.0");
    const initial = commitAll(directory, "chore: initial package");
    writeFileSync(join(directory, "packages", "example", "feature.ts"), "ok\n");
    const feature = commitAll(
      directory,
      "feat(example): add feature",
      "docs: explain feature\n\nfree-form detail",
    );

    expect(getCommitsSince(directory, "packages/example", initial)).toEqual([
      {
        hash: feature,
        author: "Test Author",
        messages: ["feat(example): add feature", "docs: explain feature"],
      },
    ]);
  });

  it("uses the first body line when a commit is not conventional", () => {
    const directory = createRepository();
    writePackage(directory, "1.0.0");
    const initial = commitAll(directory, "Initial package");
    writeFileSync(join(directory, "packages", "example", "feature.ts"), "ok\n");
    commitAll(directory, "Plain message");

    expect(
      getCommitsSince(directory, "packages/example", initial)[0],
    ).toMatchObject({
      messages: ["Plain message"],
    });
  });
});
