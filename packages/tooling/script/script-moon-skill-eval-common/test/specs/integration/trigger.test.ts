import {mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {nodeFileSystem} from "@xonovex/script-moon-common/adapters/file-system-node";
import {afterEach, describe, expect, it} from "vitest";
import {nodeProcessRunner} from "../../../src/adapters/process-runner-node.js";
import {nodeScratch} from "../../../src/adapters/scratch-node.js";
import {checkClaudeTrigger} from "../../../src/claude/trigger.js";
import {checkCodexTrigger} from "../../../src/codex/trigger.js";
import type {HarnessPorts} from "../../../src/harness-ports.js";
import {TRIGGER_OUTPUT_LIMIT} from "../../../src/trigger.js";

// These drive a real harness process through the node adapters. The scoring rules
// are covered line by line in test/specs/unit; what is left here is what only a
// spawn and a real disk reach: streamed stdout reaching the scan, an early kill
// releasing a process that would otherwise run on, and the isolated workspace a
// Codex run reads.
const ports: HarnessPorts = {
  fs: nodeFileSystem,
  scratch: nodeScratch,
  runProcess: nodeProcessRunner,
  environment: process.env,
};

const temporaryDirectories: string[] = [];

afterEach(() => {
  for (const directory of temporaryDirectories) {
    rmSync(directory, {recursive: true, force: true});
  }
  temporaryDirectories.length = 0;
});

const skillDirectory = (prefix: string, name: string): string => {
  const directory = mkdtempSync(join(tmpdir(), prefix));
  temporaryDirectories.push(directory);
  writeFileSync(
    join(directory, "SKILL.md"),
    `---\nname: ${name}\ndescription: Use for ${name} queries.\n---\n`,
  );
  return directory;
};

const runClaude = (script: string) =>
  checkClaudeTrigger(
    {
      executable: process.execPath,
      args: ["--eval", script],
      query: "test query",
      target: "plugin:test-skill",
      shortName: "test-skill",
      candidates: [{skillDirectory: "/unused", shortName: "test-skill"}],
    },
    ports,
  );

describe("checkClaudeTrigger", () => {
  it("scores a run from the lines the harness streams", async () => {
    const outcome = await runClaude(`
console.log(JSON.stringify({type: "system", subtype: "init", skills: ["plugin:test-skill"]}));
console.log(JSON.stringify({message: {content: [{type: "tool_use", name: "Skill", input: {skill: "test-skill"}}]}}));
`);

    expect(outcome).toEqual({triggered: true, error: null, selected: "target"});
  });

  it("ends a harness that would otherwise keep running once the outcome is settled", async () => {
    const outcome = await runClaude(`
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
});

describe("checkCodexTrigger", () => {
  it("stages multiple candidate skills while signaling only the target", async () => {
    const target = skillDirectory("codex-trigger-target-", "using-target");
    const competitor = skillDirectory(
      "codex-trigger-competitor-",
      "using-competitor",
    );

    const outcome = await checkCodexTrigger(
      {
        executable: process.execPath,
        args: [
          "--eval",
          `
const {existsSync, readFileSync} = require("node:fs");
const target = ".agents/skills/using-target/SKILL.md";
const competitor = ".agents/skills/using-competitor/SKILL.md";
const passed = existsSync(target) && existsSync(competitor) &&
  readFileSync(target, "utf8").includes("XONOVEX_SKILL_TRIGGERED") &&
  !readFileSync(competitor, "utf8").includes("XONOVEX_SKILL_TRIGGERED") &&
  process.env.CODEX_HOME.startsWith(process.cwd()) &&
  process.env.HOME.startsWith(process.cwd());
console.log(JSON.stringify({type:"item.completed",item:{type:"agent_message",text:passed ? "XONOVEX_SKILL_TRIGGERED" : "NOT_APPLICABLE"}}));
`,
        ],
        query: "target query",
        target: "using-target",
        shortName: "using-target",
        candidates: [
          {skillDirectory: target, shortName: "using-target"},
          {skillDirectory: competitor, shortName: "using-competitor"},
        ],
      },
      ports,
    );

    expect(outcome).toEqual({triggered: true, error: null, selected: "target"});
  });
});
