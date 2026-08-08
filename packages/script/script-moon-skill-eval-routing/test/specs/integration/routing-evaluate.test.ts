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
import {main} from "../../../src/routing-evaluate.js";

// These drive the evaluator against a harness binary on PATH. Scenario selection,
// scoring, and the failure paths are covered without a process in test/specs/unit;
// what is left here is the composition root, which resolves the harness on PATH,
// spawns it per scenario, and stages the competing guides a Codex run needs.
const addSkill = (root: string, name: string, shouldTrigger: boolean): void => {
  const plugin = join(root, `skill-${name}`);
  const guide = join(plugin, `${name}-guide`);
  mkdirSync(join(plugin, ".claude-plugin"), {recursive: true});
  mkdirSync(guide, {recursive: true});
  writeFileSync(
    join(plugin, ".claude-plugin", "plugin.json"),
    JSON.stringify({
      name: `xonovex-skill-${name}`,
      version: "0.0.0",
      skills: [`./${name}-guide`],
    }),
  );
  writeFileSync(
    join(guide, "SKILL.md"),
    `---\nname: ${name}-guide\ndescription: Use for ${name}.\n---\n`,
  );
  writeFileSync(
    join(guide, "eval-queries.json"),
    JSON.stringify([
      {
        query: "shared request",
        should_trigger: shouldTrigger,
        split: "validation",
      },
    ]),
  );
};

describe("catalog routing evaluator against a real harness", () => {
  const originalPath = process.env.PATH;
  const originalHarness = process.env.SKILL_EVAL_HARNESS;
  let directory: string;
  let catalog: string;

  beforeEach(() => {
    directory = mkdtempSync(join(tmpdir(), "routing-evaluate-"));
    catalog = join(directory, "catalog");
    mkdirSync(catalog);
    addSkill(catalog, "alpha", true);
    addSkill(catalog, "beta", false);
    process.env.PATH = `${directory}:${originalPath ?? ""}`;
    delete process.env.SKILL_EVAL_HARNESS;
    vi.spyOn(process.stderr, "write").mockImplementation(() => true);
    vi.spyOn(process.stdout, "write").mockImplementation(() => true);
  });

  afterEach(() => {
    process.env.PATH = originalPath;
    process.env.SKILL_EVAL_HARNESS = originalHarness;
    vi.restoreAllMocks();
    rmSync(directory, {recursive: true, force: true});
  });

  const executable = (name: string, script: string): void => {
    const path = join(directory, name);
    writeFileSync(path, `#!/usr/bin/env node\n${script}`);
    chmodSync(path, 0o755);
  };

  it("loads competing Claude plugins and writes passing evidence", async () => {
    executable(
      "claude",
      `
console.log(JSON.stringify({type:"system",subtype:"init",skills:["alpha-guide","beta-guide"]}));
console.log(JSON.stringify({message:{content:[{type:"tool_use",name:"Skill",input:{skill:"alpha-guide"}}]}}));
`,
    );
    const workspace = join(directory, "claude-evidence");

    const exitCode = await main([
      catalog,
      "--harness",
      "claude",
      "--workspace",
      workspace,
      "--split",
      "validation",
    ]);

    expect(exitCode).toBe(0);
    expect(
      JSON.parse(readFileSync(join(workspace, "summary.json"), "utf8")),
    ).toMatchObject({
      harness: "claude",
      catalog_scenarios: 1,
      selected_scenarios: 1,
      passed: 1,
      failed: 0,
    });
    expect(readFileSync(join(workspace, "results.jsonl"), "utf8")).toContain(
      '"candidate_skills":["alpha-guide","beta-guide"]',
    );
  });

  it("stages competing Codex guides into the run it launches", async () => {
    executable(
      "codex",
      `
const {existsSync} = require("node:fs");
const staged = existsSync(".agents/skills/alpha-guide/SKILL.md") &&
  existsSync(".agents/skills/beta-guide/SKILL.md");
console.log(JSON.stringify({type:"item.completed",item:{type:"agent_message",text: staged ? "XONOVEX_SKILL_TRIGGERED" : "NOT_APPLICABLE"}}));
`,
    );
    const workspace = join(directory, "codex-evidence");

    const exitCode = await main([
      catalog,
      "--harness=codex",
      "--workspace",
      workspace,
      "--model=gpt-test",
    ]);

    expect(exitCode).toBe(0);
    expect(
      JSON.parse(readFileSync(join(workspace, "summary.json"), "utf8")),
    ).toMatchObject({harness: "codex", passed: 1, failed: 0});
  });
});
