import {describe, expect, it} from "vitest";
import {releaseHistory, type ReleaseGit} from "../../../src/release.js";

const MANIFEST = "/repo/packages/tool/package.json";

// Answers the release reads from a recording: the hashes that touched the
// manifest, newest first, the version each one holds, and the repository tags.
const recordedGit = (recording: {
  readonly versions?: readonly (readonly [string, string])[];
  readonly tags?: readonly string[];
}): ReleaseGit => ({
  pathHistory: (path) =>
    path === "packages/tool/package.json"
      ? (recording.versions ?? []).map(([hash]) => hash)
      : [],
  readFile: (commit, path) => {
    const version = (recording.versions ?? []).find(
      ([hash]) => hash === commit,
    )?.[1];
    return path === "packages/tool/package.json" && version !== undefined
      ? JSON.stringify({name: "@xonovex/tool", version})
      : undefined;
  },
  tags: () => recording.tags ?? [],
});

describe("releaseHistory", () => {
  it("counts a package whose committed version moved as released", () => {
    const isReleased = releaseHistory(
      "/repo",
      recordedGit({
        versions: [
          ["c3", "0.2.0"],
          ["c2", "0.2.0"],
          ["c1", "0.1.0"],
        ],
      }),
    );

    expect(isReleased(MANIFEST, "0.2.0")).toBe(true);
  });

  it("counts a package with a release tag of its project as released", () => {
    const isReleased = releaseHistory(
      "/repo",
      recordedGit({versions: [["c1", "0.1.0"]], tags: ["tool-v0.1.0"]}),
    );

    expect(isReleased(MANIFEST, "0.1.0")).toBe(true);
  });

  it("counts a package that kept one version and has no tag as never released", () => {
    const isReleased = releaseHistory(
      "/repo",
      recordedGit({
        versions: [
          ["c2", "0.1.0"],
          ["c1", "0.1.0"],
        ],
        tags: [
          "v0.1.0",
          "tool-linux-x64-v0.1.0",
          "toolbox-v1.0.0",
          "tool-vnext",
          "other-v0.1.0",
        ],
      }),
    );

    expect(isReleased(MANIFEST, "0.1.0")).toBe(false);
  });

  it("counts a package with no history and no tag as never released", () => {
    const isReleased = releaseHistory("/repo", recordedGit({}));

    expect(isReleased(MANIFEST, "0.1.0")).toBe(false);
  });

  it("reads the tags once for every question", () => {
    let reads = 0;
    const git = recordedGit({});
    const isReleased = releaseHistory("/repo", {
      ...git,
      tags: () => {
        reads += 1;
        return git.tags();
      },
    });

    isReleased(MANIFEST, "0.1.0");
    isReleased("/repo/packages/other/package.json", "1.0.0");

    expect(reads).toBe(1);
  });
});
