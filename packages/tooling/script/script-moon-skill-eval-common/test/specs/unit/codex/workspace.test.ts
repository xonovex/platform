import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import {describe, expect, it} from "vitest";
import {stageCodexWorkspace} from "../../../../src/codex/workspace.js";

const sourceTree = () =>
  memoryFileSystem({
    files: {
      "/skills/alpha/SKILL.md": "# Alpha\n",
      "/skills/alpha/references/detail.md": "detail",
      "/skills/beta/SKILL.md": "# Beta\n",
    },
  });

describe("stageCodexWorkspace", () => {
  it("isolates HOME and CODEX_HOME inside the workspace", () => {
    const fs = sourceTree();

    const staged = stageCodexWorkspace("/ws", "home", [], {PATH: "/bin"}, fs);

    expect(staged).toEqual({
      cwd: "/ws",
      env: {PATH: "/bin", CODEX_HOME: "/ws/.codex", HOME: "/ws/home"},
    });
    expect(fs.isDirectory("/ws/home")).toBe(true);
    expect(fs.isDirectory("/ws/.codex")).toBe(true);
    expect(fs.isDirectory("/ws/home/.agents")).toBe(false);
  });

  it("stages project skills and appends text only where asked", () => {
    const fs = sourceTree();

    stageCodexWorkspace(
      "/ws",
      "workspace",
      [
        {source: "/skills/alpha", name: "alpha", appendix: "SIGNAL\n"},
        {source: "/skills/beta", name: "beta", appendix: undefined},
      ],
      {},
      fs,
    );

    expect(fs.readText("/ws/.agents/skills/alpha/SKILL.md")).toBe(
      "# Alpha\nSIGNAL\n",
    );
    expect(fs.readText("/ws/.agents/skills/alpha/references/detail.md")).toBe(
      "detail",
    );
    expect(fs.readText("/ws/.agents/skills/beta/SKILL.md")).toBe("# Beta\n");
  });

  it("stages user skills under the isolated home", () => {
    const fs = sourceTree();

    stageCodexWorkspace(
      "/ws",
      "home",
      [{source: "/skills/alpha", name: "target-skill", appendix: undefined}],
      {},
      fs,
    );

    expect(fs.isFile("/ws/home/.agents/skills/target-skill/SKILL.md")).toBe(
      true,
    );
  });
});
