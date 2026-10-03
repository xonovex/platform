import {describe, expect, it} from "vitest";
import {memoryFileSystem} from "../../../src/file-system-memory.js";
import {
  packageLayout,
  pluginSkillDirectories,
} from "../../../src/package-layout.js";

describe("package layout", () => {
  it("discovers direct and grouped packages without entering their fixtures", () => {
    const fs = memoryFileSystem({
      files: {
        "/repo/packages/config/config-base/package.json": "{}",
        "/repo/packages/tooling/moon/moon-nix/moon.yml": "",
        "/repo/packages/tooling/moon/moon-nix/tests/fixture/package.json": "{}",
        "/repo/packages/runtime/plugin/plugin-core/package.json": "{}",
      },
    });
    expect(packageLayout("/repo", fs)).toEqual({
      groups: [
        "/repo/packages",
        "/repo/packages/config",
        "/repo/packages/runtime",
        "/repo/packages/runtime/plugin",
        "/repo/packages/tooling",
        "/repo/packages/tooling/moon",
      ],
      packages: [
        "/repo/packages/config/config-base",
        "/repo/packages/runtime/plugin/plugin-core",
        "/repo/packages/tooling/moon/moon-nix",
      ],
    });
  });

  it("finds every skill in a grouped plugin and ignores other entries", () => {
    const fs = memoryFileSystem({
      files: {
        "/plugin/skills/alpha-guide/SKILL.md": "# Alpha",
        "/plugin/skills/beta-guide/SKILL.md": "# Beta",
        "/plugin/skills/README.md": "# Skills",
      },
    });
    expect(pluginSkillDirectories("/plugin", fs)).toEqual([
      "/plugin/skills/alpha-guide",
      "/plugin/skills/beta-guide",
    ]);
    expect(pluginSkillDirectories("/missing", fs)).toEqual([]);
  });
});
