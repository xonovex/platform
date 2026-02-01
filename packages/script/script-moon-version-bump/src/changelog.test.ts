import {describe, expect, it} from "vitest";
import {
  determineBumpLevel,
  formatCommitEntry,
  generateChangelogEntry,
} from "./changelog.js";
import type {Commit} from "./git-log.js";

describe("determineBumpLevel", () => {
  it("should detect patch bump", () => {
    expect(determineBumpLevel("1.2.3", "1.2.4")).toBe("patch");
  });

  it("should detect minor bump", () => {
    expect(determineBumpLevel("1.2.3", "1.3.0")).toBe("minor");
  });

  it("should detect major bump", () => {
    expect(determineBumpLevel("1.2.3", "2.0.0")).toBe("major");
  });
});

describe("formatCommitEntry", () => {
  it("should format a commit entry with short hash and author link", () => {
    const result = formatCommitEntry(
      "abc1234567890",
      "deorder",
      "add new feature",
    );
    expect(result).toBe(
      "- [`abc1234`](https://github.com/xonovex/platform/commit/abc1234567890) [@deorder](https://github.com/deorder)! - add new feature",
    );
  });
});

describe("generateChangelogEntry", () => {
  const commits: Commit[] = [
    {
      hash: "abc1234567890",
      author: "deorder",
      messages: ["feat: add widget"],
    },
    {
      hash: "def5678901234",
      author: "deorder",
      messages: ["fix: resolve crash"],
    },
    {
      hash: "ghi9012345678",
      author: "deorder",
      messages: ["chore: update deps"],
    },
    {
      hash: "jkl3456789012",
      author: "deorder",
      messages: ["test: add unit tests"],
    },
  ];

  it("should include feat and fix, exclude chore and test", () => {
    const entry = generateChangelogEntry("1.2.4", commits, "patch");
    expect(entry).toContain("add widget");
    expect(entry).toContain("resolve crash");
    expect(entry).not.toContain("update deps");
    expect(entry).not.toContain("add unit tests");
  });

  it("should use correct section header for patch", () => {
    const entry = generateChangelogEntry("1.2.4", commits, "patch");
    expect(entry).toContain("### Patch Changes");
  });

  it("should use correct section header for minor", () => {
    const entry = generateChangelogEntry("1.3.0", commits, "minor");
    expect(entry).toContain("### Minor Changes");
  });

  it("should use correct section header for major", () => {
    const entry = generateChangelogEntry("2.0.0", commits, "major");
    expect(entry).toContain("### Major Changes");
  });

  it("should include dependency update entries", () => {
    const entry = generateChangelogEntry("1.2.4", [], "patch", [
      {name: "@xonovex/core", version: "0.1.20"},
    ]);
    expect(entry).toContain("Updated dependency `@xonovex/core` to `0.1.20`");
  });

  it("should handle empty commits with dep updates only", () => {
    const entry = generateChangelogEntry("1.2.4", [], "patch", [
      {name: "@xonovex/core", version: "0.1.20"},
    ]);
    expect(entry).toContain("## 1.2.4");
    expect(entry).toContain("### Patch Changes");
    expect(entry).toContain("Updated dependency");
    expect(entry).not.toContain("Version bump");
  });

  it("should show version bump when no commits or dep updates", () => {
    const entry = generateChangelogEntry("1.2.4", [], "patch");
    expect(entry).toContain("## 1.2.4");
    expect(entry).toContain("### Patch Changes");
    expect(entry).toContain("- Version bump");
  });

  it("should show version bump when all commits are excluded types", () => {
    const choredCommits: Commit[] = [
      {
        hash: "abc1234567890",
        author: "deorder",
        messages: ["chore: update deps"],
      },
    ];
    const entry = generateChangelogEntry("1.2.4", choredCommits, "patch");
    expect(entry).toContain("- Version bump");
  });

  it("should handle multiple messages per commit", () => {
    const multiCommits: Commit[] = [
      {
        hash: "abc1234567890",
        author: "deorder",
        messages: [
          "feat: add logging utilities",
          "feat: add version management",
          "chore: remove deprecated scripts",
        ],
      },
    ];
    const entry = generateChangelogEntry("1.3.0", multiCommits, "minor");
    expect(entry).toContain("add logging utilities");
    expect(entry).toContain("add version management");
    expect(entry).not.toContain("remove deprecated scripts");
  });

  it("should respect custom includedTypes", () => {
    const entry = generateChangelogEntry(
      "1.2.4",
      commits,
      "patch",
      undefined,
      new Set(["chore", "test"]),
    );
    expect(entry).toContain("update deps");
    expect(entry).toContain("add unit tests");
    expect(entry).not.toContain("add widget");
    expect(entry).not.toContain("resolve crash");
  });
});
