import {
  chmodSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach, beforeEach, describe, expect, it, vi} from "vitest";
import {main} from "../../../src/evaluate.js";

// These drive the evaluator against a harness binary on PATH. Scoring, evidence,
// and the failure paths are covered without a process in test/specs/unit; what is
// left here is the composition root, which resolves the harness, probes that it
// runs, and spawns it for each query.
const QUERIES = JSON.stringify([
  {
    query: "positive",
    should_trigger: true,
    rationale: "matches",
    split: "train",
  },
  {
    query: "negative",
    should_trigger: false,
    rationale: "does not match",
    split: "validation",
  },
]);

describe("the trigger evaluator against a real harness", () => {
  const originalPath = process.env.PATH;
  let directory: string;

  beforeEach(() => {
    directory = mkdtempSync(join(tmpdir(), "skill-trigger-main-"));
    const executable = join(directory, "claude");
    writeFileSync(
      executable,
      `#!/usr/bin/env node
const args = process.argv.slice(2);
if (args.includes("--version")) {
  console.log("test claude");
  process.exit(0);
}
const query = args.at(-1) ?? "";
console.log(JSON.stringify({type: "system", subtype: "init", skills: ["test-skill"]}));
if (query === "positive") {
  console.log(JSON.stringify({message: {content: [{type: "tool_use", name: "Skill", input: {skill: "test-skill"}}]}}));
}
`,
    );
    chmodSync(executable, 0o755);
    process.env.PATH = `${directory}:${originalPath ?? ""}`;
    vi.spyOn(process.stderr, "write").mockImplementation(() => true);
    vi.spyOn(process.stdout, "write").mockImplementation(() => true);
  });

  afterEach(() => {
    process.env.PATH = originalPath;
    vi.restoreAllMocks();
    rmSync(directory, {recursive: true, force: true});
  });

  it("runs both trigger decisions and writes evidence", async () => {
    const queriesFile = join(directory, "eval-queries.json");
    const workspace = join(directory, "evidence");
    const pluginDirectory = join(directory, "plugin");
    mkdirSync(join(pluginDirectory, ".claude-plugin"), {recursive: true});
    writeFileSync(
      join(pluginDirectory, ".claude-plugin", "plugin.json"),
      JSON.stringify({
        name: "test-plugin",
        version: "0.0.0",
        dependencies: [],
        skills: ["./skills/test-skill"],
      }),
    );
    writeFileSync(queriesFile, QUERIES);

    const exitCode = await main([
      "--runs",
      "1",
      "--workspace",
      workspace,
      "--plugin-dir",
      pluginDirectory,
      queriesFile,
      "test-skill",
    ]);

    expect(exitCode).toBe(0);
    expect(
      JSON.parse(readFileSync(join(workspace, "summary.json"), "utf8")),
    ).toMatchObject({skill: "test-skill", queries: 2, passed: 2, failed: 0});
  });

  it("runs trigger decisions through the Codex harness", async () => {
    const executable = join(directory, "codex");
    const queriesFile = join(directory, "eval-queries.json");
    const workspace = join(directory, "codex-evidence");
    writeFileSync(
      join(directory, "SKILL.md"),
      "---\nname: test-skill\ndescription: Use for test queries.\n---\n",
    );
    writeFileSync(
      executable,
      String.raw`#!/usr/bin/env node
const args = process.argv.slice(2);
if (args.includes("--version")) {
  console.log("test codex");
  process.exit(0);
}
const prompt = args.at(-1) ?? "";
const text = prompt.startsWith("positive\n") ? "XONOVEX_SKILL_TRIGGERED" : "NOT_APPLICABLE";
console.log(JSON.stringify({type: "item.completed", item: {type: "agent_message", text}}));
`,
    );
    chmodSync(executable, 0o755);
    writeFileSync(queriesFile, QUERIES);

    const exitCode = await main([
      "--harness",
      "codex",
      "--runs",
      "1",
      "--workspace",
      workspace,
      queriesFile,
      "test-skill",
    ]);

    expect(exitCode).toBe(0);
    expect(
      JSON.parse(readFileSync(join(workspace, "summary.json"), "utf8")),
    ).toMatchObject({
      skill: "test-skill",
      harness: "codex",
      queries: 2,
      passed: 2,
      failed: 0,
    });
  });
});
