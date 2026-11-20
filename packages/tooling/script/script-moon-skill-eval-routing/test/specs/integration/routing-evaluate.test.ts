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
import {memoryOutput} from "@xonovex/script-moon-common/adapters/output-memory";
import {nodeEvaluatorPorts} from "@xonovex/script-moon-skill-eval-common/adapters/evaluator-ports-node";
import type {EvaluatorPorts} from "@xonovex/script-moon-skill-eval-common/harness-ports";
import {afterEach, beforeEach, describe, expect, it} from "vitest";
import {main} from "../../../src/routing-evaluate.js";

// These drive the evaluator against a harness binary on PATH. Scenario selection,
// scoring, and the failure paths are covered without a process in test/specs/unit;
// what is left here is the composition root, which resolves the harness on PATH,
// spawns it per scenario, and stages the competing skills a Codex run needs.
const addSkill = (root: string, name: string, shouldTrigger: boolean): void => {
  const plugin = join(root, `plugin-${name}`);
  const skill = join(plugin, "skills", `using-${name}`);
  mkdirSync(join(plugin, ".claude-plugin"), {recursive: true});
  mkdirSync(skill, {recursive: true});
  writeFileSync(
    join(plugin, ".claude-plugin", "plugin.json"),
    JSON.stringify({
      name: `xonovex-${name}`,
      version: "0.0.0",
      skills: "./skills/",
    }),
  );
  writeFileSync(
    join(skill, "SKILL.md"),
    `---\nname: using-${name}\ndescription: Use for ${name}.\n---\n`,
  );
  writeFileSync(
    join(skill, "eval-queries.json"),
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
  let directory: string;
  let catalog: string;
  let ports: EvaluatorPorts;

  beforeEach(() => {
    directory = mkdtempSync(join(tmpdir(), "routing-evaluate-"));
    catalog = join(directory, "catalog");
    mkdirSync(catalog);
    addSkill(catalog, "alpha", true);
    addSkill(catalog, "beta", false);
    ports = {
      ...nodeEvaluatorPorts,
      // The harness comes from each case's flag and the stub on this PATH, never
      // from the host.
      environment: {
        ...process.env,
        PATH: `${directory}:${process.env.PATH ?? ""}`,
        SKILL_EVAL_HARNESS: undefined,
      },
      output: memoryOutput(),
    };
  });

  afterEach(() => {
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
console.log(JSON.stringify({type:"system",subtype:"init",skills:["using-alpha","using-beta"]}));
console.log(JSON.stringify({message:{content:[{type:"tool_use",name:"Skill",input:{skill:"using-alpha"}}]}}));
`,
    );
    const workspace = join(directory, "claude-evidence");

    const exitCode = await main(
      [
        catalog,
        "--harness",
        "claude",
        "--workspace",
        workspace,
        "--split",
        "validation",
      ],
      directory,
      ports,
    );

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
      '"candidate_skills":["using-alpha","using-beta"]',
    );
  });

  it("stages competing Codex skills into the run it launches", async () => {
    executable(
      "codex",
      `
const {existsSync} = require("node:fs");
const staged = existsSync(".agents/skills/using-alpha/SKILL.md") &&
  existsSync(".agents/skills/using-beta/SKILL.md");
console.log(JSON.stringify({type:"item.completed",item:{type:"agent_message",text: staged ? "XONOVEX_SKILL_TRIGGERED" : "NOT_APPLICABLE"}}));
`,
    );
    const workspace = join(directory, "codex-evidence");

    const exitCode = await main(
      [
        catalog,
        "--harness=codex",
        "--workspace",
        workspace,
        "--model=gpt-test",
      ],
      directory,
      ports,
    );

    expect(exitCode).toBe(0);
    expect(
      JSON.parse(readFileSync(join(workspace, "summary.json"), "utf8")),
    ).toMatchObject({harness: "codex", passed: 1, failed: 0});
  });
});
