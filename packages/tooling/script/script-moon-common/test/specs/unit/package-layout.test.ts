import {describe, expect, it} from "vitest";
import {memoryFileSystem} from "../../../src/adapters/file-system-memory.js";
import {pluginSkillDirectories} from "../../../src/package-layout.js";

describe("package layout", () => {
  it("finds every skill in a grouped plugin and ignores other entries", () => {
    const fs = memoryFileSystem({
      files: {
        "/plugin/skills/using-alpha/SKILL.md": "# Alpha",
        "/plugin/skills/using-beta/SKILL.md": "# Beta",
        "/plugin/skills/README.md": "# Skills",
      },
    });
    expect(pluginSkillDirectories("/plugin", fs)).toEqual([
      "/plugin/skills/using-alpha",
      "/plugin/skills/using-beta",
    ]);
    expect(pluginSkillDirectories("/missing", fs)).toEqual([]);
  });
});
