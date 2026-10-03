import {describe, expect, it} from "vitest";
import {planClaudePluginStages} from "../../../src/claude-plugin.js";
import {memoryFileSystem} from "../../../src/file-system-memory.js";

describe("Claude evaluation plugins", () => {
  it("stages selected skills and dependency skills without plugin commands", () => {
    const fs = memoryFileSystem({
      files: {
        "/core/.claude-plugin/plugin.json": JSON.stringify({
          name: "xonovex-core",
          skills: "./skills/",
        }),
        "/core/commands/git-commit.md": "command",
        "/core/skills/git-guide/SKILL.md": "git",
        "/core/skills/testing-guide/SKILL.md": "testing",
        "/native/.claude-plugin/plugin.json": JSON.stringify({
          name: "xonovex-native",
          skills: "./skills/",
        }),
        "/native/skills/c99-guide/SKILL.md": "C99",
      },
    });
    expect(
      planClaudePluginStages(
        [
          "--model",
          "model",
          "--plugin-dir",
          "/native",
          "--plugin-dir",
          "/core",
        ],
        "/staged",
        ["git-guide"],
        fs,
      ),
    ).toEqual({
      args: [
        "--model",
        "model",
        "--plugin-dir",
        "/staged/plugin-3",
        "--plugin-dir",
        "/staged/plugin-5",
      ],
      stages: [
        {
          directory: "/staged/plugin-3",
          name: "xonovex-native",
          guides: ["/native/skills/c99-guide"],
        },
        {
          directory: "/staged/plugin-5",
          name: "xonovex-core",
          guides: ["/core/skills/git-guide"],
        },
      ],
    });
  });

  it("keeps standalone evaluation plugins and baseline arguments unchanged", () => {
    const fs = memoryFileSystem({
      files: {
        "/single/.claude-plugin/plugin.json": JSON.stringify({
          name: "single",
          skills: ["./guide"],
        }),
      },
    });
    expect(
      planClaudePluginStages(["--plugin-dir", "/single"], "/staged", [], fs),
    ).toEqual({args: ["--plugin-dir", "/single"], stages: []});
    expect(
      planClaudePluginStages(["--model", "model"], "/staged", [], fs),
    ).toEqual({args: ["--model", "model"], stages: []});
  });
});
