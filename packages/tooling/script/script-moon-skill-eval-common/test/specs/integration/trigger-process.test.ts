import {mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach, describe, expect, it} from "vitest";
import {
  checkCodexTriggered,
  checkTriggered,
} from "../../../src/trigger-process.js";
import {TRIGGER_OUTPUT_LIMIT} from "../../../src/trigger-scan.js";

// These drive a real harness process. The scoring rules they exercise are covered
// line by line in test/specs/unit/trigger-scan.test.ts; what is left here is the
// plumbing only a spawn can reach: streamed stdout reaching the scan, an early kill
// releasing a process that would otherwise run on, exit status and stderr arriving
// at the outcome, and the isolated workspace Codex runs in.
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

const guideDirectory = (prefix: string, name: string): string => {
  const directory = mkdtempSync(join(tmpdir(), prefix));
  temporaryDirectories.push(directory);
  writeFileSync(
    join(directory, "SKILL.md"),
    `---\nname: ${name}\ndescription: Use for ${name} queries.\n---\n`,
  );
  return directory;
};

describe("checkTriggered", () => {
  it("scores a run from the lines the harness streams", async () => {
    const outcome = await runNode(`
console.log(JSON.stringify({type: "system", subtype: "init", skills: ["plugin:test-skill"]}));
console.log(JSON.stringify({message: {content: [{type: "tool_use", name: "Skill", input: {skill: "test-skill"}}]}}));
`);

    expect(outcome).toEqual({triggered: true, error: null, selected: "target"});
  });

  it("ends a harness that would otherwise keep running once the outcome is settled", async () => {
    const outcome = await runNode(`
console.log(JSON.stringify({type: "system", subtype: "init", skills: ["test-skill"]}));
console.log(JSON.stringify({type: "stream_event", event: {type: "content_block_delta", delta: {type: "text_delta", text: "x".repeat(${String(TRIGGER_OUTPUT_LIMIT + 1)})}}}));
setInterval(() => {}, 1000);
`);

    expect(outcome).toEqual({
      triggered: false,
      error: null,
      selected: "output-limit",
    });
  });

  it("carries the exit status and stderr of a failed harness into the outcome", async () => {
    const outcome = await runNode(
      `process.stderr.write("authentication failed"); process.exit(3);`,
    );

    expect(outcome).toEqual({
      triggered: false,
      error: "claude exited 3: authentication failed",
    });
  });

  it("reports a harness that cannot be spawned", async () => {
    const outcome = await checkTriggered(
      "test query",
      [],
      "plugin:test-skill",
      "test-skill",
      join(tmpdir(), "xonovex-missing-harness"),
    );

    expect(outcome.triggered).toBe(false);
    expect(outcome.error).toContain("ENOENT");
  });
});

describe("checkCodexTriggered", () => {
  it("stages multiple candidate guides while signaling only the target", async () => {
    const target = guideDirectory("codex-trigger-target-", "target-guide");
    const competitor = guideDirectory(
      "codex-trigger-competitor-",
      "competitor-guide",
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
  !readFileSync(competitor, "utf8").includes("XONOVEX_SKILL_TRIGGERED") &&
  process.env.CODEX_HOME.startsWith(process.cwd()) &&
  process.env.HOME.startsWith(process.cwd());
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

    expect(outcome).toEqual({triggered: true, error: null, selected: "target"});
  });

  it("carries the exit status and stderr of a failed harness into the outcome", async () => {
    const outcome = await checkCodexTriggered({
      args: [
        "--eval",
        'process.stderr.write("authentication failed"); process.exit(3);',
      ],
      executable: process.execPath,
      guideDirectory: guideDirectory("codex-trigger-guide-", "test-skill"),
      query: "test query",
      shortName: "test-skill",
    });

    expect(outcome).toEqual({
      triggered: false,
      error: "codex exited 3: authentication failed",
    });
  });
});
