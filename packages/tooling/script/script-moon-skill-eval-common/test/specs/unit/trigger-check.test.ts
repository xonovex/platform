import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import {describe, expect, it} from "vitest";
import {memoryProcessRunner} from "../../../src/adapters/process-runner-memory.js";
import {memoryScratch} from "../../../src/adapters/scratch-memory.js";
import {checkTrigger} from "../../../src/trigger-check.js";

describe("checkTrigger", () => {
  it.each([
    ["claude", "xonovex-claude-trigger-1"],
    ["codex", "xonovex-codex-trigger-1"],
  ] as const)(
    "runs the %s harness in its own workspace",
    async (harness, workspace) => {
      const fs = memoryFileSystem({files: {"/skill/SKILL.md": "# Skill\n"}});
      const scratch = memoryScratch(fs);

      await checkTrigger(
        harness,
        {
          executable: `/bin/${harness}`,
          args: [],
          query: "q",
          target: "skill",
          shortName: "skill",
          candidates: [{skillDirectory: "/skill", shortName: "skill"}],
        },
        {
          fs,
          scratch,
          runProcess: memoryProcessRunner(() => ({})).run,
          environment: {},
        },
      );

      expect(scratch.removed()).toEqual([`/scratch/${workspace}`]);
    },
  );
});
