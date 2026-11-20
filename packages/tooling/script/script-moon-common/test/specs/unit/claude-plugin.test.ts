import {describe, expect, it} from "vitest";
import {memoryFileSystem} from "../../../src/adapters/file-system-memory.js";
import {
  planClaudePluginStages,
  stageClaudePluginArguments,
} from "../../../src/claude-plugin.js";

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

  it("writes each staged plugin with only the selected skills and their resources", () => {
    const fs = memoryFileSystem({
      files: {
        "/core/.claude-plugin/plugin.json": JSON.stringify({
          name: "xonovex-core",
          skills: "./skills/",
        }),
        "/core/skills/using-git/SKILL.md": "# Git",
        "/core/skills/using-git/references/commits.md":
          "Use conventional commits.",
        "/core/skills/authoring-tests/SKILL.md": "# Testing",
      },
    });

    const args = stageClaudePluginArguments(
      ["--plugin-dir", "/core"],
      "/staged",
      ["using-git"],
      fs,
    );

    expect(args).toEqual(["--plugin-dir", "/staged/plugin-1"]);
    expect(
      fs.readText("/staged/plugin-1/skills/using-git/references/commits.md"),
    ).toBe("Use conventional commits.");
    expect(fs.isDirectory("/staged/plugin-1/skills/authoring-tests")).toBe(
      false,
    );
    expect(
      JSON.parse(fs.readText("/staged/plugin-1/.claude-plugin/plugin.json")),
    ).toEqual({name: "xonovex-core", skills: ["./skills/using-git"]});
  });
});
