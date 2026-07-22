import {mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach, describe, expect, it} from "vitest";
import {
  checkCodexTriggered,
  checkTriggered,
  TRIGGER_OUTPUT_LIMIT,
} from "./trigger-process.js";

const temporaryDirectories: string[] = [];

afterEach(() => {
  for (const directory of temporaryDirectories) {
    rmSync(directory, {recursive: true, force: true});
  }
  temporaryDirectories.length = 0;
});

const runNode = (script: string) =>
  checkTriggered(
    "test query",
    ["--eval", script],
    "plugin:test-skill",
    "test-skill",
    process.execPath,
  );

const runCodexNode = (script: string) => {
  const guideDirectory = mkdtempSync(join(tmpdir(), "codex-trigger-guide-"));
  temporaryDirectories.push(guideDirectory);
  writeFileSync(
    join(guideDirectory, "SKILL.md"),
    "---\nname: test-skill\ndescription: Use for test queries.\n---\n",
  );
  return checkCodexTriggered({
    args: ["--eval", script],
    executable: process.execPath,
    guideDirectory,
    query: "test query",
    shortName: "test-skill",
  });
};

describe("checkTriggered", () => {
  it("detects a matching Skill tool invocation", async () => {
    const outcome = await runNode(`
console.log(JSON.stringify({type: "system", subtype: "init", skills: ["plugin:test-skill"]}));
console.log(JSON.stringify({message: {content: [{type: "tool_use", name: "Skill", input: {skill: "test-skill"}}]}}));
`);

    expect(outcome).toEqual({triggered: true, error: null});
  });

  it("returns a clean negative result when the skill is available", async () => {
    const outcome = await runNode(`
console.log(JSON.stringify({type: "system", subtype: "init", skills: ["test-skill"]}));
console.log("not json");
`);

    expect(outcome).toEqual({triggered: false, error: null});
  });

  it("reports when the target skill is unavailable", async () => {
    const outcome = await runNode(
      `console.log(JSON.stringify({type: "system", subtype: "init", skills: ["other-skill"]}));`,
    );

    expect(outcome).toEqual({
      triggered: false,
      error: "target skill unavailable",
    });
  });

  it("includes process stderr in nonzero-exit failures", async () => {
    const outcome = await runNode(
      `process.stderr.write("authentication failed"); process.exit(3);`,
    );

    expect(outcome).toEqual({
      triggered: false,
      error: "claude exited 3: authentication failed",
    });
  });

  it("stops a response that exceeds the output limit", async () => {
    const outcome = await runNode(`
console.log(JSON.stringify({type: "system", subtype: "init", skills: ["test-skill"]}));
console.log(JSON.stringify({type: "stream_event", event: {type: "content_block_delta", delta: {type: "text_delta", text: "x".repeat(${String(TRIGGER_OUTPUT_LIMIT + 1)})}}}));
setInterval(() => {}, 1000);
`);

    expect(outcome).toEqual({triggered: false, error: "output-limit"});
  });
});

describe("checkCodexTriggered", () => {
  it("detects the isolated skill signal in a Codex JSONL agent message", async () => {
    const outcome = await runCodexNode(
      'console.log(JSON.stringify({type:"item.completed",item:{type:"agent_message",text:"XONOVEX_SKILL_TRIGGERED"}}))',
    );

    expect(outcome).toEqual({triggered: true, error: null});
  });

  it("returns a clean negative result for malformed and nonmatching output", async () => {
    const outcome = await runCodexNode(`
console.log("not json");
console.log(JSON.stringify({type: "item.completed", item: {type: "agent_message", text: "NOT_APPLICABLE"}}));
`);

    expect(outcome).toEqual({triggered: false, error: null});
  });

  it("includes process stderr in nonzero-exit failures", async () => {
    const outcome = await runCodexNode(
      'process.stderr.write("authentication failed"); process.exit(3);',
    );

    expect(outcome).toEqual({
      triggered: false,
      error: "codex exited 3: authentication failed",
    });
  });

  it("stops a response that exceeds the output limit", async () => {
    const outcome = await runCodexNode(`
console.log(JSON.stringify({type: "item.completed", item: {type: "agent_message", text: "x".repeat(${String(TRIGGER_OUTPUT_LIMIT + 1)})}}));
setInterval(() => {}, 1000);
`);

    expect(outcome).toEqual({triggered: false, error: "output-limit"});
  });
});
