import {describe, expect, it} from "vitest";
import {
  findChangelogRange,
  getCommitsSince,
  isIncludedType,
  parseConventionalCommit,
} from "../../../src/git-log.js";
import type {VersionGit} from "../../../src/git.js";

const PKG = "packages/core";

// Answers the history reads these functions make from a recording rather than a
// repository. `absent` names the hashes at which the manifest does not exist,
// which is how a path introduced later looks.
const recordedGit = (recording: {
  readonly hashes?: readonly string[];
  readonly versions?: Readonly<Record<string, string>>;
  readonly absent?: readonly string[];
  readonly parents?: readonly string[];
  readonly commits?: VersionGit["commitsSince"];
}): VersionGit => ({
  pathHistory: () => recording.hashes ?? [],
  readFile: (commit) =>
    (recording.absent ?? []).includes(commit)
      ? undefined
      : JSON.stringify({
          name: "@xonovex/core",
          version: recording.versions?.[commit] ?? "0.0.0",
        }),
  hasParent: (commit) => (recording.parents ?? []).includes(commit),
  commitsSince: recording.commits ?? (() => []),
});

describe("parseConventionalCommit", () => {
  it("reads a type and description, with and without a scope", () => {
    expect(parseConventionalCommit("feat: add a thing")).toEqual({
      type: "feat",
      description: "add a thing",
      breaking: false,
    });
    expect(parseConventionalCommit("fix(core): repair a thing")).toEqual({
      type: "fix",
      description: "repair a thing",
      breaking: false,
    });
  });

  it("marks a subject with the breaking-change marker, with and without a scope", () => {
    expect(parseConventionalCommit("feat!: drop a flag")).toEqual({
      type: "feat",
      description: "drop a flag",
      breaking: true,
    });
    expect(parseConventionalCommit("refactor(core)!: rename a port")).toEqual({
      type: "refactor",
      description: "rename a port",
      breaking: true,
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

    expect(findChangelogRange(PKG, "1.2.3", git)).toEqual({since: "c2"});
  });

  it("skips a commit where the manifest does not yet exist", () => {
    const git = recordedGit({
      hashes: ["c3", "c2", "c1"],
      versions: {c3: "1.2.3", c1: "1.0.0"},
      absent: ["c2"],
    });

    expect(findChangelogRange(PKG, "1.2.3", git)).toEqual({since: "c1"});
  });

  it("returns undefined when the package has no history at all", () => {
    expect(findChangelogRange(PKG, "1.2.3", recordedGit({}))).toBeUndefined();
  });

  it("covers from before the introducing commit when no version ever differed", () => {
    const git = recordedGit({
      hashes: ["c2", "c1"],
      versions: {c2: "1.2.3", c1: "1.2.3"},
      parents: ["c1"],
    });

    expect(findChangelogRange(PKG, "1.2.3", git)).toEqual({
      since: "c1~1",
    });
  });

  it("covers the whole history when the introducing commit is the root", () => {
    const git = recordedGit({
      hashes: ["c1"],
      versions: {c1: "1.2.3"},
    });

    expect(findChangelogRange(PKG, "1.2.3", git)).toEqual({
      since: undefined,
    });
  });
});

describe("getCommitsSince", () => {
  it("keeps the conventional lines of each commit message", () => {
    const git = recordedGit({
      commits: () => [
        {
          hash: "aaa",
          author: "Ada",
          message: "feat: add a thing\n\nfix: and repair another",
        },
        {hash: "bbb", author: "Grace", message: "refactor: tidy up"},
      ],
    });

    expect(getCommitsSince(PKG, "c1", git)).toEqual([
      {
        hash: "aaa",
        author: "Ada",
        messages: ["feat: add a thing", "fix: and repair another"],
      },
      {hash: "bbb", author: "Grace", messages: ["refactor: tidy up"]},
    ]);
  });

  it("keeps a breaking-change subject as a conventional line", () => {
    const git = recordedGit({
      commits: () => [
        {
          hash: "aaa",
          author: "Ada",
          message: "feat(core)!: drop a flag\n\nbody",
        },
      ],
    });

    expect(getCommitsSince(PKG, "c1", git)).toEqual([
      {hash: "aaa", author: "Ada", messages: ["feat(core)!: drop a flag"]},
    ]);
  });

  it("falls back to the first line when no line is conventional", () => {
    const git = recordedGit({
      commits: () => [
        {hash: "aaa", author: "Ada", message: "merge branch main\nnoise"},
        {hash: "bbb", author: "Grace", message: ""},
      ],
    });

    expect(getCommitsSince(PKG, "c1", git)).toEqual([
      {hash: "aaa", author: "Ada", messages: ["merge branch main"]},
      {hash: "bbb", author: "Grace", messages: []},
    ]);
  });

  it("asks for the history of the package since the given ref", () => {
    const seen: [string | undefined, string][] = [];
    const git = recordedGit({
      commits: (since, path) => {
        seen.push([since, path]);
        return [];
      },
    });

    getCommitsSince(PKG, undefined, git);
    getCommitsSince(PKG, "c1", git);

    expect(seen).toEqual([
      [undefined, PKG],
      ["c1", PKG],
    ]);
  });
});
