import {mkdirSync, mkdtempSync, rmSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach, beforeEach, describe, expect, it} from "vitest";
import {findWorkspaceRoot} from "./workspace.js";

describe("findWorkspaceRoot", () => {
  let tmp: string;

  beforeEach(() => {
    tmp = mkdtempSync(join(tmpdir(), "moon-scripts-common-"));
  });

  afterEach(() => {
    rmSync(tmp, {recursive: true, force: true});
  });

  it("should find the directory containing .moon", () => {
    mkdirSync(join(tmp, ".moon"));
    const nested = join(tmp, "a", "b", "c");
    mkdirSync(nested, {recursive: true});

    expect(findWorkspaceRoot(nested)).toBe(tmp);
  });

  it("should return start dir if it contains .moon", () => {
    mkdirSync(join(tmp, ".moon"));
    expect(findWorkspaceRoot(tmp)).toBe(tmp);
  });

  it("should throw if no .moon directory is found", () => {
    expect(() => findWorkspaceRoot(tmp)).toThrow(
      "Could not find workspace root",
    );
  });
});
