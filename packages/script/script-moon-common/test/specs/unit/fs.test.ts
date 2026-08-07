import {mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach, beforeEach, describe, expect, it} from "vitest";
import {
  isDirectory,
  isFile,
  resolveClaudePluginDirectories,
  resolveGuideDirectory,
} from "../../../src/fs.js";

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

  it("resolves local Claude plugin dependencies before the target", () => {
    const dependency = join(root, "dependency");
    const target = join(root, "target");
    for (const directory of [dependency, target]) {
      mkdirSync(join(directory, ".claude-plugin"), {recursive: true});
    }
    writeFileSync(
      join(dependency, ".claude-plugin", "plugin.json"),
      JSON.stringify({name: "dependency", skills: ["./guide"]}),
    );
    writeFileSync(
      join(target, ".claude-plugin", "plugin.json"),
      JSON.stringify({
        name: "target",
        skills: ["./guide"],
        dependencies: ["dependency"],
      }),
    );

    expect(resolveClaudePluginDirectories(target)).toEqual([
      dependency,
      target,
    ]);
  });

  it("rejects a missing local Claude plugin dependency", () => {
    const target = join(root, "target");
    mkdirSync(join(target, ".claude-plugin"), {recursive: true});
    writeFileSync(
      join(target, ".claude-plugin", "plugin.json"),
      JSON.stringify({
        name: "target",
        skills: ["./guide"],
        dependencies: ["missing"],
      }),
    );

    expect(() => resolveClaudePluginDirectories(target)).toThrow(
      "target -> missing",
    );
  });

  it("rejects non-skill Claude plugin capabilities", () => {
    const target = join(root, "target");
    mkdirSync(join(target, ".claude-plugin"), {recursive: true});
    writeFileSync(
      join(target, ".claude-plugin", "plugin.json"),
      JSON.stringify({name: "target", skills: ["./guide"], hooks: {}}),
    );
    expect(() => resolveClaudePluginDirectories(target)).toThrow(
      "is not skill-only",
    );

    writeFileSync(
      join(target, ".claude-plugin", "plugin.json"),
      JSON.stringify({name: "target", skills: ["./guide"]}),
    );
    mkdirSync(join(target, "hooks"));
    expect(() => resolveClaudePluginDirectories(target)).toThrow(
      "contains hooks",
    );
  });
});
