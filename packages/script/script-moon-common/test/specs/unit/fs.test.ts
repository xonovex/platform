import {join} from "node:path";
import {describe, expect, it} from "vitest";
import {
  memoryFileSystem,
  type MemoryTree,
} from "../../../src/file-system-memory.js";
import {
  isDirectory,
  isFile,
  resolveClaudePluginDirectories,
  resolveGuideDirectory,
} from "../../../src/fs.js";

const ROOT = "/catalog";

const manifest = (
  name: string,
  extra: Readonly<Record<string, unknown>> = {},
): string => JSON.stringify({name, skills: ["./guide"], ...extra});

const pluginPath = (directory: string): string =>
  join(directory, ".claude-plugin", "plugin.json");

const tree = (value: MemoryTree) => memoryFileSystem(value);

describe("filesystem helpers", () => {
  it("distinguishes files, directories, and missing paths", () => {
    const fs = tree({files: {[join(ROOT, "file.txt")]: "data"}});

    expect(isFile(join(ROOT, "file.txt"), fs)).toBe(true);
    expect(isDirectory(ROOT, fs)).toBe(true);
    expect(isFile(join(ROOT, "missing"), fs)).toBe(false);
    expect(isDirectory(join(ROOT, "missing"), fs)).toBe(false);
    expect(isFile(ROOT, fs)).toBe(false);
  });

  it("resolves a guide at the base", () => {
    const fs = tree({files: {[join(ROOT, "SKILL.md")]: "guide"}});

    expect(resolveGuideDirectory(ROOT, fs)).toBe(ROOT);
  });

  it("resolves a guide one level below the base", () => {
    const fs = tree({files: {[join(ROOT, "guide", "SKILL.md")]: "guide"}});

    expect(resolveGuideDirectory(ROOT, fs)).toBe(join(ROOT, "guide"));
  });

  it("falls back to the base when nothing below it holds a guide", () => {
    const fs = tree({directories: [join(ROOT, "empty")]});

    expect(resolveGuideDirectory(ROOT, fs)).toBe(ROOT);
  });

  it("rejects an ambiguous package", () => {
    const fs = tree({
      files: {
        [join(ROOT, "one", "SKILL.md")]: "guide",
        [join(ROOT, "two", "SKILL.md")]: "guide",
      },
    });

    expect(() => resolveGuideDirectory(ROOT, fs)).toThrow("multiple SKILL.md");
  });

  it("resolves local Claude plugin dependencies before the target", () => {
    const dependency = join(ROOT, "dependency");
    const target = join(ROOT, "target");
    const fs = tree({
      files: {
        [pluginPath(dependency)]: manifest("dependency"),
        [pluginPath(target)]: manifest("target", {
          dependencies: ["dependency"],
        }),
      },
    });

    expect(resolveClaudePluginDirectories(target, fs)).toEqual([
      dependency,
      target,
    ]);
  });

  it("visits a shared dependency once", () => {
    const shared = join(ROOT, "shared");
    const middle = join(ROOT, "middle");
    const target = join(ROOT, "target");
    const fs = tree({
      files: {
        [pluginPath(shared)]: manifest("shared"),
        [pluginPath(middle)]: manifest("middle", {dependencies: ["shared"]}),
        [pluginPath(target)]: manifest("target", {
          dependencies: ["shared", "middle"],
        }),
      },
    });

    expect(resolveClaudePluginDirectories(target, fs)).toEqual([
      shared,
      middle,
      target,
    ]);
  });

  it("rejects a missing local Claude plugin dependency", () => {
    const target = join(ROOT, "target");
    const fs = tree({
      files: {
        [pluginPath(target)]: manifest("target", {dependencies: ["missing"]}),
      },
    });

    expect(() => resolveClaudePluginDirectories(target, fs)).toThrow(
      "target -> missing",
    );
  });

  it("rejects a non-skill capability declared in the manifest", () => {
    const target = join(ROOT, "target");
    const fs = tree({
      files: {[pluginPath(target)]: manifest("target", {hooks: {}})},
    });

    expect(() => resolveClaudePluginDirectories(target, fs)).toThrow(
      "is not skill-only",
    );
  });

  it("rejects a non-skill component present on disk", () => {
    const target = join(ROOT, "target");
    const fs = tree({
      files: {[pluginPath(target)]: manifest("target")},
      directories: [join(target, "hooks")],
    });

    expect(() => resolveClaudePluginDirectories(target, fs)).toThrow(
      "contains hooks",
    );
  });

  it("rejects a manifest that is malformed, unnamed, or wrongly shaped", () => {
    const target = join(ROOT, "target");
    for (const [contents, message] of [
      ["{", "invalid Claude plugin manifest"],
      ["[]", "invalid Claude plugin manifest"],
      [JSON.stringify({skills: ["./guide"]}), "has no name"],
      [
        JSON.stringify({
          name: "target",
          skills: ["./guide"],
          dependencies: [1],
        }),
        "dependencies are invalid",
      ],
      [JSON.stringify({name: "target", skills: []}), "skills are invalid"],
    ] as const) {
      const fs = tree({files: {[pluginPath(target)]: contents}});

      expect(() => resolveClaudePluginDirectories(target, fs)).toThrow(message);
    }
  });

  it("ignores a sibling directory that carries no plugin manifest", () => {
    const target = join(ROOT, "target");
    const fs = tree({
      files: {
        [pluginPath(target)]: manifest("target"),
        [join(ROOT, "not-a-plugin", "README.md")]: "prose",
      },
    });

    expect(resolveClaudePluginDirectories(target, fs)).toEqual([target]);
  });
});
