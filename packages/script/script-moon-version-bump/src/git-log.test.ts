import {execFileSync} from "node:child_process";
import {mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";
import {afterEach, describe, expect, it} from "vitest";
import {
  getCommitsSince,
  getLastVersionRef,
  isIncludedType,
  parseConventionalCommit,
} from "./git-log.js";

const directories: string[] = [];
const gitExecutable = resolveExecutable("git");

const createRepository = (): string => {
  const directory = mkdtempSync(join(tmpdir(), "version-git-log-"));
  directories.push(directory);
  execFileSync(gitExecutable, ["init", "--quiet"], {cwd: directory});
  execFileSync(gitExecutable, ["config", "user.name", "Test Author"], {
    cwd: directory,
  });
  execFileSync(gitExecutable, ["config", "user.email", "test@example.com"], {
    cwd: directory,
  });
  mkdirSync(join(directory, "packages", "example"), {recursive: true});
  return directory;
};

const writePackage = (directory: string, version: string): void => {
  writeFileSync(
    join(directory, "packages", "example", "package.json"),
    `${JSON.stringify({name: "example", version}, null, 2)}\n`,
  );
};

const commit = (directory: string, subject: string, body?: string): string => {
  execFileSync(gitExecutable, ["add", "."], {cwd: directory});
  execFileSync(
    gitExecutable,
    ["commit", "--quiet", "-m", subject, ...(body ? ["-m", body] : [])],
    {cwd: directory},
  );
  return execFileSync(gitExecutable, ["rev-parse", "HEAD"], {
    cwd: directory,
    encoding: "utf8",
  }).trim();
};

afterEach(() => {
  for (const directory of directories) {
    rmSync(directory, {recursive: true, force: true});
  }
  directories.length = 0;
});

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
    const previousVersionCommit = commit(directory, "feat: initial package");
    writePackage(directory, "2.0.0");
    commit(directory, "chore: release 2.0.0");

    expect(getLastVersionRef(directory, "packages/example", "2.0.0")).toBe(
      previousVersionCommit,
    );
  });

  it("falls back to before the package introduction when its version never changed", () => {
    const directory = createRepository();
    writePackage(directory, "1.0.0");
    const introduction = commit(directory, "feat: initial package");

    expect(getLastVersionRef(directory, "packages/example", "1.0.0")).toBe(
      `${introduction}~1`,
    );
    expect(
      getLastVersionRef(directory, "packages/missing", "1.0.0"),
    ).toBeUndefined();
  });

  it("extracts conventional subject and body messages from package commits", () => {
    const directory = createRepository();
    writePackage(directory, "1.0.0");
    const initial = commit(directory, "chore: initial package");
    writeFileSync(join(directory, "packages", "example", "feature.ts"), "ok\n");
    const feature = commit(
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
    const initial = commit(directory, "Initial package");
    writeFileSync(join(directory, "packages", "example", "feature.ts"), "ok\n");
    commit(directory, "Plain message");

    expect(
      getCommitsSince(directory, "packages/example", initial)[0],
    ).toMatchObject({
      messages: ["Plain message"],
    });
  });
});
