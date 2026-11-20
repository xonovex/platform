import {join} from "node:path";
import {describe, expect, it} from "vitest";
import {
  memoryFileSystem,
  type MemoryTree,
} from "../../../src/adapters/file-system-memory.js";
import {
  findPluginDirectory,
  resolveSkillDirectory,
} from "../../../src/skill-directory.js";

const ROOT = "/catalog";

const manifest = (
  name: string,
  extra: Readonly<Record<string, unknown>> = {},
): string => JSON.stringify({name, skills: ["./guide"], ...extra});

const pluginPath = (directory: string): string =>
  join(directory, ".claude-plugin", "plugin.json");

const tree = (value: MemoryTree) => memoryFileSystem(value);

describe("skill directories", () => {
  it("finds a skill's grouped plugin", () => {
    const plugin = join(ROOT, "plugin-core");
    const skill = join(plugin, "skills", "using-git");
    const fs = tree({
      files: {
        [pluginPath(plugin)]: manifest("xonovex-core", {skills: "./skills/"}),
        [join(skill, "SKILL.md")]: "skill",
      },
    });
    expect(findPluginDirectory(skill, fs)).toBe(plugin);
    expect(findPluginDirectory(join(plugin, "guide"), fs)).toBe(plugin);
  });
  it("resolves a skill at the base", () => {
    const fs = tree({files: {[join(ROOT, "SKILL.md")]: "guide"}});

    expect(resolveSkillDirectory(ROOT, fs)).toBe(ROOT);
  });

  it("resolves a skill one level below the base", () => {
    const fs = tree({files: {[join(ROOT, "guide", "SKILL.md")]: "guide"}});

    expect(resolveSkillDirectory(ROOT, fs)).toBe(join(ROOT, "guide"));
  });

  it("falls back to the base when nothing below it holds a skill", () => {
    const fs = tree({directories: [join(ROOT, "empty")]});

    expect(resolveSkillDirectory(ROOT, fs)).toBe(ROOT);
  });

  it("rejects an ambiguous package", () => {
    const fs = tree({
      files: {
        [join(ROOT, "one", "SKILL.md")]: "guide",
        [join(ROOT, "two", "SKILL.md")]: "guide",
      },
    });

    expect(() => resolveSkillDirectory(ROOT, fs)).toThrow("multiple SKILL.md");
  });
});
