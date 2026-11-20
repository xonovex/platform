import {describe, expect, it} from "vitest";
import {memoryFileSystem} from "../../../src/adapters/file-system-memory.js";
import {pluginHandoffNames} from "../../../src/skill-handoffs.js";

describe("pluginHandoffNames", () => {
  it("reads the bold skill names of each SKILL.md and reference", () => {
    const fs = memoryFileSystem({
      files: {
        "/plugin/skills/using-alpha/SKILL.md":
          "# Alpha\n\nCommit with **using-git**, not using-docker.\n",
        "/plugin/skills/using-alpha/references/push.md":
          "# Push\n\nOpen it with **using-github**.\n",
        "/plugin/skills/using-alpha/references/notes.txt": "**using-zod**\n",
        "/plugin/skills/using-beta/SKILL.md": "# Beta\n\nSee `applying-tdd`.\n",
      },
    });

    expect([...pluginHandoffNames("/plugin", fs)].toSorted()).toEqual([
      "using-git",
      "using-github",
    ]);
  });

  it("returns no name for a plugin without skills", () => {
    const fs = memoryFileSystem({files: {"/plugin/README.md": "# Plugin\n"}});

    expect(pluginHandoffNames("/plugin", fs).size).toBe(0);
  });
});
