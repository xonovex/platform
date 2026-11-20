import {join} from "node:path";
import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import type {GitRepository} from "@xonovex/script-moon-common/git";
import {describe, expect, it} from "vitest";
import {detectVersionChanges} from "../../../src/detect.js";

const ROOT = "/repo";

const manifestAt = (source: string, manifest: unknown): [string, string] => [
  join(ROOT, source, "package.json"),
  JSON.stringify(manifest),
];

const detectManifest = (manifest: unknown): readonly string[] =>
  detectVersionChanges(
    ROOT,
    "commit",
    [{id: "package", source: "package"}],
    {readFile: () => void 0},
    memoryFileSystem({
      files: Object.fromEntries([manifestAt("package", manifest)]),
    }),
  );

describe("detectVersionChanges", () => {
  it("reports version changes and skips packages absent at the ref", () => {
    const fs = memoryFileSystem({
      files: Object.fromEntries([
        manifestAt("packages/changed", {name: "changed", version: "2.0.0"}),
        manifestAt("packages/added", {name: "added", version: "1.0.0"}),
      ]),
    });
    const git: Pick<GitRepository, "readFile"> = {
      readFile: (_commit, path) =>
        path.includes("changed")
          ? JSON.stringify({name: "changed", version: "1.0.0"})
          : undefined,
    };

    expect(
      detectVersionChanges(
        ROOT,
        "commit",
        [
          {id: "changed", source: "packages/changed"},
          {id: "added", source: "packages/added"},
        ],
        git,
        fs,
      ),
    ).toEqual(["changed"]);
  });

  it("skips a project whose manifest is absent from the working tree", () => {
    const fs = memoryFileSystem({directories: [ROOT]});

    expect(
      detectVersionChanges(
        ROOT,
        "commit",
        [{id: "gone", source: "packages/gone"}],
        {readFile: () => JSON.stringify({name: "gone", version: "1.0.0"})},
        fs,
      ),
    ).toEqual([]);
  });

  it("reports nothing when the version is unchanged at the ref", () => {
    const fs = memoryFileSystem({
      files: Object.fromEntries([
        manifestAt("packages/same", {name: "same", version: "1.0.0"}),
      ]),
    });

    expect(
      detectVersionChanges(
        ROOT,
        "commit",
        [{id: "same", source: "packages/same"}],
        {readFile: () => JSON.stringify({name: "same", version: "1.0.0"})},
        fs,
      ),
    ).toEqual([]);
  });

  it("rejects malformed publishable package metadata", () => {
    const fs = memoryFileSystem({
      files: Object.fromEntries([manifestAt("package", {name: "package"})]),
    });

    expect(() =>
      detectVersionChanges(
        ROOT,
        "commit",
        [{id: "package", source: "package"}],
        {readFile: () => void 0},
        fs,
      ),
    ).toThrow("non-empty package version");
  });

  it("rejects a manifest without a name or that is not an object", () => {
    expect(() => detectManifest({version: "1.0.0"})).toThrow(
      "non-empty package name",
    );
    expect(() => detectManifest([])).toThrow("must contain a JSON object");
  });

  it("reports a package that was private at the ref and is publishable now", () => {
    const fs = memoryFileSystem({
      files: Object.fromEntries([
        manifestAt("packages/opened", {name: "opened", version: "1.0.0"}),
        manifestAt("packages/closed", {
          name: "closed",
          version: "1.0.0",
          private: true,
        }),
      ]),
    });

    expect(
      detectVersionChanges(
        ROOT,
        "commit",
        [
          {id: "opened", source: "packages/opened"},
          {id: "closed", source: "packages/closed"},
        ],
        {
          readFile: () =>
            JSON.stringify({name: "opened", version: "1.0.0", private: true}),
        },
        fs,
      ),
    ).toEqual(["opened"]);
  });
});
