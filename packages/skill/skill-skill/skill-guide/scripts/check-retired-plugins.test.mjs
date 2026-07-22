import {mkdirSync, mkdtempSync, rmSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach, describe, expect, it} from "vitest";
import {findRetired} from "./check-retired-plugins.mjs";

const temporaryDirectories = [];

afterEach(() => {
  for (const directory of temporaryDirectories) {
    rmSync(directory, {recursive: true, force: true});
  }
  temporaryDirectories.length = 0;
});

describe("findRetired", () => {
  it("reports installed versions with their replacement", () => {
    const root = mkdtempSync(join(tmpdir(), "retired-plugin-"));
    temporaryDirectories.push(root);
    mkdirSync(join(root, "xonovex-skill-prompt", "4.0.0"), {recursive: true});

    expect(findRetired([root])).toEqual([
      expect.objectContaining({
        retired: "xonovex-skill-prompt",
        replacement: "xonovex-skill-command",
        version: "4.0.0",
      }),
    ]);
  });
});
