import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import {describe, expect, it} from "vitest";
import {readWorkspacePackages} from "../../../src/workspace-packages.js";
import {emptyHistory, manifestsAtHead} from "../../util/version-git.js";

const manifestPath = (directory: string): string =>
  `/repo/packages/${directory}/package.json`;

describe("readWorkspacePackages", () => {
  it("reads each manifest from the tree and its version at HEAD", () => {
    const fs = memoryFileSystem({
      files: {
        [manifestPath("base")]: JSON.stringify({
          name: "@xonovex/base",
          version: "0.2.0",
        }),
        [manifestPath("new")]: JSON.stringify({
          name: "@xonovex/new",
          version: "0.0.1",
        }),
      },
    });
    const git = emptyHistory({
      readFile: manifestsAtHead({
        "packages/base/package.json": {version: "0.1.0"},
      }),
    });

    expect(
      readWorkspacePackages(
        "/repo",
        git,
        [manifestPath("base"), manifestPath("new")],
        fs,
      ),
    ).toEqual([
      {
        path: manifestPath("base"),
        pkg: {name: "@xonovex/base", version: "0.2.0"},
        headVersion: "0.1.0",
      },
      {
        path: manifestPath("new"),
        pkg: {name: "@xonovex/new", version: "0.0.1"},
        headVersion: undefined,
      },
    ]);
  });
});
