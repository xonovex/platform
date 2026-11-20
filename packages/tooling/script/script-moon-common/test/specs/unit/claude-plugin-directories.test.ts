import {join} from "node:path";
import {describe, expect, it} from "vitest";
import {
  memoryFileSystem,
  type MemoryTree,
} from "../../../src/adapters/file-system-memory.js";
import {resolveClaudePluginDirectories} from "../../../src/claude-plugin.js";

const ROOT = "/catalog";

const manifest = (
  name: string,
  extra: Readonly<Record<string, unknown>> = {},
): string => JSON.stringify({name, skills: ["./guide"], ...extra});

const pluginPath = (directory: string): string =>
  join(directory, ".claude-plugin", "plugin.json");

const tree = (value: MemoryTree) => memoryFileSystem(value);

describe("resolveClaudePluginDirectories", () => {
  it("resolves a grouped plugin to itself", () => {
    const plugin = join(ROOT, "plugin-core");
    const fs = tree({
      files: {
        [pluginPath(plugin)]: manifest("xonovex-core", {skills: "./skills/"}),
        [join(plugin, "skills", "using-git", "SKILL.md")]: "skill",
      },
    });
    expect(resolveClaudePluginDirectories(plugin, fs)).toEqual([plugin]);
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

  it("rejects a plugin that holds a commands directory", () => {
    const plugin = join(ROOT, "plugin-core");
    const fs = tree({
      files: {
        [pluginPath(plugin)]: manifest("xonovex-core", {skills: "./skills/"}),
        [join(plugin, "commands", "run.md")]: "command",
      },
    });

    expect(() => resolveClaudePluginDirectories(plugin, fs)).toThrow(
      "contains commands",
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
      [JSON.stringify({name: "target", skills: [42]}), "skills are invalid"],
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
