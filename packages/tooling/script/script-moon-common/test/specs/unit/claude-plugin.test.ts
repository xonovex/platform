import {describe, expect, it} from "vitest";
import {planClaudePluginStages} from "../../../src/claude-plugin.js";
import {memoryFileSystem} from "../../../src/file-system-memory.js";

describe("Claude evaluation plugins", () => {
  it("stages selected skills and dependency skills without the other skills of the plugin", () => {
    const fs = memoryFileSystem({
      files: {
        "/core/.claude-plugin/plugin.json": JSON.stringify({
          name: "xonovex-core",
          skills: "./skills/",
        }),
        "/core/skills/using-git/SKILL.md": "git",
        "/core/skills/authoring-tests/SKILL.md": "testing",
        "/native/.claude-plugin/plugin.json": JSON.stringify({
          name: "xonovex-native",
          skills: "./skills/",
        }),
        "/native/skills/writing-c99/SKILL.md": "C99",
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
        ["using-git"],
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
          skillDirectories: ["/native/skills/writing-c99"],
        },
        {
          directory: "/staged/plugin-5",
          name: "xonovex-core",
          skillDirectories: ["/core/skills/using-git"],
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
