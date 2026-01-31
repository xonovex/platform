import {describe, expect, it} from "vitest";
import {isIncludedType, parseConventionalCommit} from "./git-log.js";

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
});
