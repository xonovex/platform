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

  it("scores a competing skill invocation as a non-trigger", async () => {
    const outcome = await runNode(`
console.log(JSON.stringify({type: "system", subtype: "init", skills: ["plugin:test-skill"]}));
console.log(JSON.stringify({message: {content: [{type: "tool_use", name: "Skill", input: {skill: "other-skill"}}]}}));
process.exit(1);
`);

    expect(outcome).toEqual({triggered: false, error: null});
  });

  it("scores a denied competing skill invocation as a non-trigger", async () => {
    const outcome = await runNode(`
console.log(JSON.stringify({type: "system", subtype: "init", skills: ["plugin:test-skill"]}));
console.log(JSON.stringify({permission_denials: [{tool_name: "Skill", tool_input: {skill: "other-skill"}}]}));
process.exit(1);
`);

    expect(outcome).toEqual({triggered: false, error: null});
  });

  it("prefers the target when it is invoked alongside a competing skill", async () => {
    const outcome = await runNode(`
console.log(JSON.stringify({type: "system", subtype: "init", skills: ["plugin:test-skill"]}));
console.log(JSON.stringify({message: {content: [{type: "tool_use", name: "Skill", input: {skill: "other-skill"}}, {type: "tool_use", name: "Skill", input: {skill: "test-skill"}}]}}));
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

  it("scores a response that exceeds the output limit as a non-trigger", async () => {
    const outcome = await runNode(`
console.log(JSON.stringify({type: "system", subtype: "init", skills: ["test-skill"]}));
console.log(JSON.stringify({type: "stream_event", event: {type: "content_block_delta", delta: {type: "text_delta", text: "x".repeat(${String(TRIGGER_OUTPUT_LIMIT + 1)})}}}));
setInterval(() => {}, 1000);
`);

    expect(outcome).toEqual({triggered: false, error: null});
  });

  it("still reports an unavailable skill when the output limit is exceeded", async () => {
    const outcome = await runNode(`
console.log(JSON.stringify({type: "system", subtype: "init", skills: ["other-skill"]}));
console.log(JSON.stringify({type: "stream_event", event: {type: "content_block_delta", delta: {type: "text_delta", text: "x".repeat(${String(TRIGGER_OUTPUT_LIMIT + 1)})}}}));
setInterval(() => {}, 1000);
`);

    expect(outcome).toEqual({
      triggered: false,
      error: "target skill unavailable",
    });
  });
});

describe("checkCodexTriggered", () => {
  it("detects the isolated skill signal in a Codex JSONL agent message", async () => {
    const outcome = await runCodexNode(
      'console.log(JSON.stringify({type:"item.completed",item:{type:"agent_message",text:"XONOVEX_SKILL_TRIGGERED"}}))',
    );

    expect(outcome).toEqual({triggered: true, error: null});
  });

  it("stages multiple candidate guides while signaling only the target", async () => {
    const target = mkdtempSync(join(tmpdir(), "codex-trigger-target-"));
    const competitor = mkdtempSync(join(tmpdir(), "codex-trigger-competitor-"));
    temporaryDirectories.push(target, competitor);
    writeFileSync(
      join(target, "SKILL.md"),
      "---\nname: target-guide\ndescription: Use for targets.\n---\n",
    );
    writeFileSync(
      join(competitor, "SKILL.md"),
      "---\nname: competitor-guide\ndescription: Use for competitors.\n---\n",
    );

    const outcome = await checkCodexTriggered({
      args: [
        "--eval",
        `
const {existsSync, readFileSync} = require("node:fs");
const target = ".agents/skills/target-guide/SKILL.md";
const competitor = ".agents/skills/competitor-guide/SKILL.md";
const passed = existsSync(target) && existsSync(competitor) &&
  readFileSync(target, "utf8").includes("XONOVEX_SKILL_TRIGGERED") &&
  !readFileSync(competitor, "utf8").includes("XONOVEX_SKILL_TRIGGERED");
console.log(JSON.stringify({type:"item.completed",item:{type:"agent_message",text:passed ? "XONOVEX_SKILL_TRIGGERED" : "NOT_APPLICABLE"}}));
`,
      ],
      executable: process.execPath,
      guideDirectory: target,
      query: "target query",
      shortName: "target-guide",
      candidateGuides: [
        {guideDirectory: target, shortName: "target-guide"},
        {guideDirectory: competitor, shortName: "competitor-guide"},
      ],
    });

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

  it("scores a response that exceeds the output limit as a non-trigger", async () => {
    const outcome = await runCodexNode(`
console.log(JSON.stringify({type: "item.completed", item: {type: "agent_message", text: "x".repeat(${String(TRIGGER_OUTPUT_LIMIT + 1)})}}));
setInterval(() => {}, 1000);
`);

    expect(outcome).toEqual({triggered: false, error: null});
  });
});
