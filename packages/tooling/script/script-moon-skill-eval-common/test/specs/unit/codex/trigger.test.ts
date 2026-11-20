import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import {describe, expect, it} from "vitest";
import {memoryProcessRunner} from "../../../../src/adapters/process-runner-memory.js";
import {memoryScratch} from "../../../../src/adapters/scratch-memory.js";
import {CODEX_TRIGGER_SIGNAL} from "../../../../src/codex/trigger-scan.js";
import {
  checkCodexTrigger,
  codexTriggerPrompt,
} from "../../../../src/codex/trigger.js";
import type {TriggerRequest} from "../../../../src/trigger.js";

const REQUEST: TriggerRequest = {
  executable: "/bin/codex",
  args: ["exec", "--json"],
  query: "the query",
  target: "using-target",
  shortName: "using-target",
  candidates: [
    {skillDirectory: "/skills/target", shortName: "using-target"},
    {skillDirectory: "/skills/competitor", shortName: "using-competitor"},
  ],
};

const message = (text: string): string =>
  JSON.stringify({type: "item.completed", item: {type: "agent_message", text}});

const setup = (stdout: string) => {
  const fs = memoryFileSystem({
    files: {
      "/skills/target/SKILL.md": "# Target\n",
      "/skills/competitor/SKILL.md": "# Competitor\n",
    },
  });
  const runner = memoryProcessRunner(() => ({stdout}));
  const scratch = memoryScratch(fs);
  const outcome = checkCodexTrigger(REQUEST, {
    fs,
    scratch,
    runProcess: runner.run,
    environment: {PATH: "/bin"},
  });
  return {fs, runner, scratch, outcome};
};

describe("checkCodexTrigger", () => {
  it("scores the signal in the agent message as a trigger", async () => {
    const {outcome} = setup(message(CODEX_TRIGGER_SIGNAL));

    await expect(outcome).resolves.toEqual({
      triggered: true,
      error: null,
      selected: "target",
    });
  });

  it("stages every candidate and signals only the target", async () => {
    const {fs, outcome, scratch} = setup(message("NOT_APPLICABLE"));
    await outcome;

    const skills = "/scratch/xonovex-codex-trigger-1/.agents/skills";
    expect(fs.readText(`${skills}/using-target/SKILL.md`)).toContain(
      CODEX_TRIGGER_SIGNAL,
    );
    expect(fs.readText(`${skills}/using-competitor/SKILL.md`)).toBe(
      "# Competitor\n",
    );
    expect(scratch.removed()).toEqual(["/scratch/xonovex-codex-trigger-1"]);
  });

  it("starts Codex in the workspace with an isolated home and the prompt last", async () => {
    const {outcome, runner} = setup(message("NOT_APPLICABLE"));
    await expect(outcome).resolves.toMatchObject({selected: "none"});

    const workspace = "/scratch/xonovex-codex-trigger-1";
    expect(runner.invocations()).toEqual([
      {
        executable: "/bin/codex",
        args: ["exec", "--json", codexTriggerPrompt("the query")],
        cwd: workspace,
        env: {
          PATH: "/bin",
          CODEX_HOME: `${workspace}/.codex`,
          HOME: `${workspace}/home`,
        },
        timeoutMs: 60_000,
      },
    ]);
  });
});
