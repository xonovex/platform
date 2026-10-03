import {
  existsSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import {tmpdir} from "node:os";
import {dirname, join} from "node:path";
import {describe, expect, it} from "vitest";
import {stageClaudePluginArguments} from "../../../src/claude-plugin.js";

describe("Claude plugin staging", () => {
  it("copies the selected skill and its resources without commands or other skills", () => {
    const root = mkdtempSync(join(tmpdir(), "moon-common-claude-plugin-"));
    const plugin = join(root, "source");
    const write = (path: string, content: string): void => {
      const target = join(plugin, path);
      mkdirSync(dirname(target), {recursive: true});
      writeFileSync(target, content);
    };
    try {
      write(
        ".claude-plugin/plugin.json",
        JSON.stringify({name: "xonovex-core", skills: "./skills/"}),
      );
      write("commands/git-commit.md", "Commit the change.");
      write("skills/git-guide/SKILL.md", "# Git");
      write(
        "skills/git-guide/references/commits.md",
        "Use conventional commits.",
      );
      write("skills/testing-guide/SKILL.md", "# Testing");

      const args = stageClaudePluginArguments(
        ["--plugin-dir", plugin],
        join(root, "staged"),
        ["git-guide"],
      );
      const staged = args[1];
      if (staged === undefined)
        throw new Error("missing staged plugin argument");

      expect(
        readFileSync(
          join(staged, "skills/git-guide/references/commits.md"),
          "utf8",
        ),
      ).toBe("Use conventional commits.");
      expect(existsSync(join(staged, "commands"))).toBe(false);
      expect(existsSync(join(staged, "skills/testing-guide"))).toBe(false);
      expect(
        JSON.parse(
          readFileSync(join(staged, ".claude-plugin/plugin.json"), "utf8"),
        ),
      ).toEqual({name: "xonovex-core", skills: ["./skills/git-guide"]});
    } finally {
      rmSync(root, {recursive: true, force: true});
    }
  });
});
