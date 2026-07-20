import {describe, expect, it} from "vitest";
import {checkTriggered, TRIGGER_OUTPUT_LIMIT} from "./trigger-process.js";

const runNode = (script: string) =>
  checkTriggered(
    "test query",
    ["--eval", script],
    "plugin:test-skill",
    "test-skill",
    process.execPath,
  );

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
