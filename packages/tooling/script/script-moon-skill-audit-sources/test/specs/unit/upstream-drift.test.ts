import {describe, expect, it} from "vitest";
import {parseSources} from "../../../src/source-parser.js";
import {computeDrift} from "../../../src/upstream-drift.js";
import type {UpstreamRepository} from "../../../src/upstream.js";

const source = (fields: string) => {
  const [parsed] = parseSources(
    `# Sources\n\n## Library\n- **Provenance:** Upstream library\n${fields}`,
  );
  if (parsed === undefined) throw new Error("fixture has no source");
  return parsed;
};

const repository = (
  overrides: Partial<UpstreamRepository> = {},
): UpstreamRepository => ({
  fetchTags: () => true,
  tags: () => ["v1.0.0", "notes", "1.2.0", "v1.1.9"],
  hasCommit: () => true,
  commits: () => [],
  changedFiles: () => [],
  ...overrides,
});

describe("computeDrift", () => {
  it("reports a checkout that is not a repository", () => {
    expect(
      computeDrift("/up", source("- **Checkout:** up\n"), false),
    ).toMatchObject({
      resolved: false,
      note: "checkout not found or not a git repo",
    });
  });

  it("reports a failed tag fetch", () => {
    expect(
      computeDrift(
        "/up",
        source("- **Checkout:** up\n"),
        true,
        repository({fetchTags: () => false}),
      ),
    ).toMatchObject({
      resolved: true,
      pull_failed: true,
      note: "failed to fetch upstream tags",
    });
  });

  it("needs a pinned commit or version that exists", () => {
    expect(
      computeDrift(
        "/up",
        source("- **Checkout:** up\n"),
        false,
        repository({tags: () => []}),
      ),
    ).toMatchObject({
      latest_version: null,
      note: "no **Commit:** or **Version:** to diff from",
    });
    expect(
      computeDrift(
        "/up",
        source("- **Checkout:** up\n- **Commit:** abcdef1\n"),
        false,
        repository({hasCommit: () => false}),
      ),
    ).toMatchObject({note: "pinned ref 'abcdef1' not found in checkout"});
  });

  it("lists the commits since the pin and the references their files feed", () => {
    const commits = Array.from({length: 25}, (_, index) => `c${String(index)}`);
    const ranges: string[] = [];

    const drift = computeDrift(
      "/up",
      source(
        "- **Checkout:** up\n- **Version:** 1.1.0\n- **Watch:** src/api -> api.md\n- **Watch:** docs -> guide.md, api.md\n",
      ),
      false,
      repository({
        commits: (range) => {
          ranges.push(range);
          return commits;
        },
        changedFiles: () => ["src/api/x.ts", "README.md"],
      }),
    );

    expect(ranges).toEqual(["1.1.0..1.2.0"]);
    expect(drift).toMatchObject({
      latest_version: "1.2.0",
      behind: true,
      commit_count: 25,
      review_refs: ["api.md"],
    });
    expect(drift.commits).toHaveLength(20);
  });
});
