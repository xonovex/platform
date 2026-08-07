import {mkdirSync, mkdtempSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {describe, expect, it} from "vitest";
import {detectVersionChanges, type GitReader} from "../../../src/detect.js";

describe("detectVersionChanges", () => {
  it("reports version changes and skips packages absent at the ref", () => {
    const root = mkdtempSync(join(tmpdir(), "version-detect-"));
    mkdirSync(join(root, "packages", "changed"), {recursive: true});
    mkdirSync(join(root, "packages", "added"), {recursive: true});
    writeFileSync(
      join(root, "packages", "changed", "package.json"),
      JSON.stringify({name: "changed", version: "2.0.0"}),
    );
    writeFileSync(
      join(root, "packages", "added", "package.json"),
      JSON.stringify({name: "added", version: "1.0.0"}),
    );
    const git: GitReader = {
      readFile: (_commit, path) =>
        path.includes("changed")
          ? JSON.stringify({name: "changed", version: "1.0.0"})
          : undefined,
    };

    expect(
      detectVersionChanges(
        root,
        "commit",
        [
          {id: "changed", source: "packages/changed"},
          {id: "added", source: "packages/added"},
        ],
        git,
      ),
    ).toEqual(["changed"]);
  });

  it("rejects malformed publishable package metadata", () => {
    const root = mkdtempSync(join(tmpdir(), "version-detect-"));
    mkdirSync(join(root, "package"), {recursive: true});
    writeFileSync(
      join(root, "package", "package.json"),
      JSON.stringify({name: "package"}),
    );

    expect(() =>
      detectVersionChanges(
        root,
        "commit",
        [{id: "package", source: "package"}],
        {readFile: () => void 0},
      ),
    ).toThrow("non-empty package version");
  });
});
