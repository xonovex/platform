import {describe, expect, it} from "vitest";
import {parseCommitLog, parseGitLines} from "../../../src/git.js";

describe("parseCommitLog", () => {
  it("splits records into hash, author, and the full message", () => {
    const output =
      "\0abc123|Ada Lovelace\nfeat: add engine\n\nBody line.\n\n" +
      "\0def456|Alan Turing\nfix(core): repair\n";

    expect(parseCommitLog(output)).toEqual([
      {
        hash: "abc123",
        author: "Ada Lovelace",
        message: "feat: add engine\n\nBody line.",
      },
      {hash: "def456", author: "Alan Turing", message: "fix(core): repair"},
    ]);
  });

  it("returns no records for an empty log", () => {
    expect(parseCommitLog("")).toEqual([]);
  });
});

describe("parseGitLines", () => {
  it("drops the empty lines of a listing", () => {
    expect(parseGitLines("a.txt\nnested/\n\n")).toEqual(["a.txt", "nested/"]);
  });
});
