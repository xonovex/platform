import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import {afterEach, beforeEach, describe, expect, it, vi} from "vitest";
import {
  planChangelog,
  serializePackage,
  type ChangelogRequest,
} from "../../../src/package-changes.js";
import {emptyHistory} from "../../util/version-git.js";

const PACKAGE_PATH = "/repo/packages/core/package.json";
const CHANGELOG_PATH = "/repo/packages/core/CHANGELOG.md";

const request = (
  overrides: Partial<ChangelogRequest> = {},
): ChangelogRequest => ({
  rootDir: "/repo",
  packagePath: PACKAGE_PATH,
  packageName: "@xonovex/core",
  oldVersion: "1.0.0",
  newVersion: "1.1.0",
  dryRun: false,
  depUpdates: [],
  changelogFilename: undefined,
  gitBase: "base",
  includedTypes: undefined,
  git: emptyHistory({
    commitsSince: () => [
      {hash: "abc1234", author: "Ada", message: "feat: add a thing"},
    ],
  }),
  fs: memoryFileSystem(),
  ...overrides,
});

beforeEach(() => {
  vi.spyOn(console, "error").mockImplementation(() => {});
  vi.spyOn(console, "log").mockImplementation(() => {});
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe("serializePackage", () => {
  it("writes indented JSON with a final newline", () => {
    expect(serializePackage({name: "a"})).toBe('{\n  "name": "a"\n}\n');
  });
});

describe("planChangelog", () => {
  it("creates a changelog from the commits since the base", () => {
    const since: (string | undefined)[] = [];
    const change = planChangelog(
      request({
        git: emptyHistory({
          commitsSince: (ref, path) => {
            since.push(ref, path);
            return [
              {hash: "abc1234", author: "Ada", message: "feat: add a thing"},
            ];
          },
        }),
      }),
    );

    expect(since).toEqual(["base", "packages/core"]);
    expect(change?.path).toBe(CHANGELOG_PATH);
    expect(change?.content).toContain("# @xonovex/core");
    expect(change?.content).toContain("## 1.1.0");
    expect(change?.content).toContain("add a thing");
  });

  it("adds the entry above the content of an existing changelog", () => {
    const fs = memoryFileSystem({
      files: {[CHANGELOG_PATH]: "# @xonovex/core\n\n## 1.0.0\n\n- First\n"},
    });

    const change = planChangelog(request({fs}));

    expect(change?.content.indexOf("## 1.1.0")).toBeLessThan(
      change?.content.indexOf("## 1.0.0") ?? 0,
    );
    expect(change?.content).toContain("- First");
  });

  it("writes to the requested file name", () => {
    expect(
      planChangelog(request({changelogFilename: "HISTORY.md"}))?.path,
    ).toBe("/repo/packages/core/HISTORY.md");
  });

  it("finds the range from the history when no base is given", () => {
    const change = planChangelog(
      request({
        gitBase: undefined,
        git: emptyHistory({
          pathHistory: () => ["c2"],
          readFile: () => JSON.stringify({version: "0.9.0"}),
        }),
      }),
    );

    expect(change?.content).toContain("## 1.1.0");
  });

  it("skips a package with no history", () => {
    expect(planChangelog(request({gitBase: undefined}))).toBeUndefined();
  });

  it("prints the entry and plans no change on a dry run", () => {
    const log = vi.spyOn(console, "log").mockImplementation(() => {});

    expect(planChangelog(request({dryRun: true}))).toBeUndefined();
    expect(log).toHaveBeenCalledWith(expect.stringContaining("## 1.1.0"));
  });
});
