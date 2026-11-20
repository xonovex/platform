import {join} from "node:path";
import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import {describe, expect, it} from "vitest";
import {
  memoryProcessRunner,
  type ScriptedProcess,
} from "../../../../src/adapters/process-runner-memory.js";
import {memoryScratch} from "../../../../src/adapters/scratch-memory.js";
import {checkClaudeTrigger} from "../../../../src/claude/trigger.js";
import type {TriggerRequest} from "../../../../src/trigger.js";

const PLUGIN = "/catalog/plugin-test";

const pluginTree = (
  manifest = '{"name":"xonovex-test","skills":"./skills/"}',
) =>
  memoryFileSystem({
    files: {
      [join(PLUGIN, ".claude-plugin", "plugin.json")]: manifest,
      [join(PLUGIN, "skills", "test-skill", "SKILL.md")]: "# Test\n",
      [join(PLUGIN, "skills", "other-skill", "SKILL.md")]: "# Other\n",
    },
  });

const REQUEST: TriggerRequest = {
  executable: "/bin/claude",
  args: ["-p", "--plugin-dir", PLUGIN],
  query: "the query",
  target: "plugin:test-skill",
  shortName: "test-skill",
  candidates: [{skillDirectory: "/unused", shortName: "test-skill"}],
};

const SKILL_CALL = [
  JSON.stringify({type: "system", subtype: "init", skills: ["test-skill"]}),
  JSON.stringify({
    message: {
      content: [
        {type: "tool_use", name: "Skill", input: {skill: "test-skill"}},
      ],
    },
  }),
].join("\n");

const check = (fs: FileSystem, answer: ScriptedProcess) => {
  const runner = memoryProcessRunner(() => answer);
  const scratch = memoryScratch(fs);
  const outcome = checkClaudeTrigger(REQUEST, {
    fs,
    scratch,
    runProcess: runner.run,
    environment: {PATH: "/bin"},
  });
  return {outcome, runner, scratch};
};

describe("checkClaudeTrigger", () => {
  it("scores the streamed skill call as a trigger", async () => {
    const {outcome} = check(pluginTree(), {stdout: SKILL_CALL});

    await expect(outcome).resolves.toEqual({
      triggered: true,
      error: null,
      selected: "target",
    });
  });

  it("stages only the candidate skills and passes the query last", async () => {
    const fs = pluginTree();
    const {outcome, runner, scratch} = check(fs, {stdout: SKILL_CALL});
    await outcome;

    const workspace = "/scratch/xonovex-claude-trigger-1";
    const staged = join(workspace, "plugin-2");
    expect(runner.invocations()).toEqual([
      {
        executable: "/bin/claude",
        args: ["-p", "--plugin-dir", staged, "the query"],
        cwd: undefined,
        env: {PATH: "/bin"},
        timeoutMs: 60_000,
      },
    ]);
    expect(
      JSON.parse(fs.readText(join(staged, ".claude-plugin", "plugin.json"))),
    ).toEqual({name: "xonovex-test", skills: ["./skills/test-skill"]});
    expect(scratch.removed()).toEqual([workspace]);
  });

  it("reports a failed exit as an infrastructure error", async () => {
    const {outcome} = check(pluginTree(), {code: 3, stderr: "denied"});

    await expect(outcome).resolves.toEqual({
      triggered: false,
      error: "claude exited 3: denied",
    });
  });

  it("removes the workspace when staging fails", async () => {
    const {outcome, scratch} = check(pluginTree("not json"), {});

    await expect(outcome).rejects.toThrow();
    expect(scratch.removed()).toEqual(["/scratch/xonovex-claude-trigger-1"]);
  });
});
