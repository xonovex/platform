import {mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach, beforeEach, describe, expect, it} from "vitest";
import {isDirectory, isFile, resolveGuideDirectory} from "./fs.js";

describe("filesystem helpers", () => {
  let root: string;

  beforeEach(() => {
    root = mkdtempSync(join(tmpdir(), "moon-common-fs-"));
  });

  afterEach(() => {
    rmSync(root, {recursive: true, force: true});
  });

  it("distinguishes files, directories, and missing paths", () => {
    const file = join(root, "file.txt");
    writeFileSync(file, "data");
    expect(isFile(file)).toBe(true);
    expect(isDirectory(root)).toBe(true);
    expect(isFile(join(root, "missing"))).toBe(false);
  });

  it("resolves a guide at the base or one level below it", () => {
    writeFileSync(join(root, "SKILL.md"), "guide");
    expect(resolveGuideDirectory(root)).toBe(root);

    rmSync(join(root, "SKILL.md"));
    const guide = join(root, "guide");
    mkdirSync(guide);
    writeFileSync(join(guide, "SKILL.md"), "guide");
    expect(resolveGuideDirectory(root)).toBe(guide);
  });

  it("rejects an ambiguous package", () => {
    for (const name of ["one", "two"]) {
      mkdirSync(join(root, name));
      writeFileSync(join(root, name, "SKILL.md"), "guide");
    }
    expect(() => resolveGuideDirectory(root)).toThrow("multiple SKILL.md");
  });
});
