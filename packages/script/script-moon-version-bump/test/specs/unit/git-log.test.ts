import {describe, expect, it} from "vitest";
import {
  findChangelogRange,
  getCommitsSince,
  isIncludedType,
  parseConventionalCommit,
} from "../../../src/git-log.js";
import type {GitRunner} from "../../../src/git.js";

const ROOT = "/workspace";
const PKG = "packages/core";

// Answers the four git reads these functions make, from recorded output rather
// than a repository. `absent` names the hashes at which the manifest does not
// exist, which is how ls-tree reports a path introduced later.
const recordedGit = (recording: {
  readonly hashes?: readonly string[];
  readonly versions?: Readonly<Record<string, string>>;
  readonly absent?: readonly string[];
  readonly parents?: Readonly<Record<string, readonly string[]>>;
  readonly log?: string;
}): GitRunner => {
  return (args) => {
    const [command] = args;
    if (command === "log" && args[1] === "--format=%H") {
      return `${(recording.hashes ?? []).join("\n")}\n`;
    }
    if (command === "log") return recording.log ?? "";
    if (command === "ls-tree") {
      const hash = args[3] ?? "";
      return (recording.absent ?? []).includes(hash)
        ? ""
        : `${PKG}/package.json\0`;
    }
    if (command === "show") {
      const hash = (args[1] ?? "").split(":", 1)[0] ?? "";
      return `${JSON.stringify({
        name: "@xonovex/core",
        version: recording.versions?.[hash] ?? "0.0.0",
      })}\n`;
    }
    if (command === "rev-list") {
      const hash = args.at(-1) ?? "";
      return [hash, ...(recording.parents?.[hash] ?? [])].join(" ");
    }
    throw new Error(`unexpected git command ${String(command)}`);
  };
};

describe("parseConventionalCommit", () => {
  it("reads a type and description, with and without a scope", () => {
    expect(parseConventionalCommit("feat: add a thing")).toEqual({
      type: "feat",
      description: "add a thing",
    });
    expect(parseConventionalCommit("fix(core): repair a thing")).toEqual({
      type: "fix",
      description: "repair a thing",
    });
  });

  it("returns undefined for a non-conventional subject", () => {
    expect(parseConventionalCommit("just some words")).toBeUndefined();
  });
});

describe("isIncludedType", () => {
  it("uses the default set when none is given and honours an override", () => {
    expect(isIncludedType("feat")).toBe(true);
    expect(isIncludedType("chore")).toBe(false);
    expect(isIncludedType("chore", new Set(["chore"]))).toBe(true);
  });
});

describe("findChangelogRange", () => {
  it("bounds the release at the first commit whose version differs", () => {
    const git = recordedGit({
      hashes: ["c3", "c2", "c1"],
      versions: {c3: "1.2.3", c2: "1.1.0", c1: "1.0.0"},
    });

    expect(findChangelogRange(ROOT, PKG, "1.2.3", git)).toEqual({since: "c2"});
  });

  it("skips a commit where the manifest does not yet exist", () => {
    const git = recordedGit({
      hashes: ["c3", "c2", "c1"],
      versions: {c3: "1.2.3", c1: "1.0.0"},
      absent: ["c2"],
    });

    expect(findChangelogRange(ROOT, PKG, "1.2.3", git)).toEqual({since: "c1"});
  });

  it("returns undefined when the package has no history at all", () => {
    expect(
      findChangelogRange(ROOT, PKG, "1.2.3", recordedGit({})),
    ).toBeUndefined();
  });

  it("covers from before the introducing commit when no version ever differed", () => {
    const git = recordedGit({
      hashes: ["c2", "c1"],
      versions: {c2: "1.2.3", c1: "1.2.3"},
      parents: {c1: ["c0"]},
    });

    expect(findChangelogRange(ROOT, PKG, "1.2.3", git)).toEqual({
      since: "c1~1",
    });
  });

  it("covers the whole history when the introducing commit is the root", () => {
    const git = recordedGit({
      hashes: ["c1"],
      versions: {c1: "1.2.3"},
      parents: {},
    });

    expect(findChangelogRange(ROOT, PKG, "1.2.3", git)).toEqual({
      since: undefined,
    });
  });
});

describe("getCommitsSince", () => {
  it("reads the hash, author and conventional subjects of each commit", () => {
    const git = recordedGit({
      log: "\0aaa|Ada\nfeat: add a thing\n\nfix: and repair another\0bbb|Grace\nrefactor: tidy up\n",
    });

    expect(getCommitsSince(ROOT, PKG, "c1", git)).toEqual([
      {
        hash: "aaa",
        author: "Ada",
        messages: ["feat: add a thing", "fix: and repair another"],
      },
      {hash: "bbb", author: "Grace", messages: ["refactor: tidy up"]},
    ]);
  });

  it("falls back to the first body line when no subject is conventional", () => {
    const git = recordedGit({log: "\0aaa|Ada\nmerge branch main\nnoise\n"});

    expect(getCommitsSince(ROOT, PKG, "c1", git)).toEqual([
      {hash: "aaa", author: "Ada", messages: ["merge branch main"]},
    ]);
  });

  it("reads the whole history when no starting ref is given", () => {
    const seen: string[][] = [];
    const git: GitRunner = (args) => {
      seen.push([...args]);
      return "";
    };

    getCommitsSince(ROOT, PKG, undefined, git);

    expect(seen[0]).toContain("HEAD");
    expect(seen[0]?.some((arg) => arg.includes(".."))).toBe(false);
  });
});
