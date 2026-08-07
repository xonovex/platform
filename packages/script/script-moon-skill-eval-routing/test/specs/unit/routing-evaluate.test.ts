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

describe("catalog routing evaluator", () => {
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
      model: "claude-haiku-4-5-20251001",
      catalog_scenarios: 1,
      selected_scenarios: 1,
      passed: 1,
      failed: 0,
    });
    expect(readFileSync(join(workspace, "results.jsonl"), "utf8")).toContain(
      '"candidate_skills":["alpha-guide","beta-guide"]',
    );
  });

  it("stages competing Codex guides and honors limit and offset", async () => {
    executable(
      "codex",
      'console.log(JSON.stringify({type:"item.completed",item:{type:"agent_message",text:"XONOVEX_SKILL_TRIGGERED"}}));',
    );
    const workspace = join(directory, "codex-evidence");

    const exitCode = await main([
      catalog,
      "--harness=codex",
      "--workspace",
      workspace,
      "--offset=0",
      "--limit=1",
      "--model=gpt-test",
    ]);

    expect(exitCode).toBe(0);
    expect(
      JSON.parse(readFileSync(join(workspace, "summary.json"), "utf8")),
    ).toMatchObject({
      harness: "codex",
      model: "gpt-test",
      selected_scenarios: 1,
    });
  });

  it("filters routing scenarios to selected expected owners", async () => {
    executable(
      "claude",
      `
console.log(JSON.stringify({type:"system",subtype:"init",skills:["alpha-guide","beta-guide"]}));
console.log(JSON.stringify({message:{content:[{type:"tool_use",name:"Skill",input:{skill:"alpha-guide"}}]}}));
`,
    );
    const workspace = join(directory, "owner-evidence");

    const exitCode = await main([
      catalog,
      "--workspace",
      workspace,
      "--owners",
      "alpha-guide",
    ]);

    expect(exitCode).toBe(0);
    expect(
      JSON.parse(readFileSync(join(workspace, "summary.json"), "utf8")),
    ).toMatchObject({
      selected_owners: ["alpha-guide"],
      selected_scenarios: 1,
    });
    await expect(main([catalog, "--owners", "beta-guide"])).rejects.toThrow(
      "no routing scenarios selected",
    );
  });

  it("returns one when a healthy run selects no owner", async () => {
    executable(
      "claude",
      'console.log(JSON.stringify({type:"system",subtype:"init",skills:["alpha-guide","beta-guide"]}));',
    );

    await expect(
      main([catalog, "--workspace", join(directory, "failed-evidence")]),
    ).resolves.toBe(1);
  });

  it("returns two and records an infrastructure failure", async () => {
    executable(
      "claude",
      'console.log(JSON.stringify({type:"system",subtype:"init",skills:["beta-guide"]}));',
    );
    const workspace = join(directory, "invalid-evidence");

    await expect(main([catalog, "--workspace", workspace])).resolves.toBe(2);
    expect(readFileSync(join(workspace, "invalid-run.json"), "utf8")).toContain(
      "target skill unavailable",
    );
  });

  it("rejects invalid selection and evaluator options", async () => {
    await expect(main([catalog, "--harness", "other"])).rejects.toThrow(
      "invalid harness",
    );
    await expect(main([catalog, "--split", "other"])).rejects.toThrow(
      "invalid split",
    );
    await expect(main([catalog, "--runs", "4"])).rejects.toThrow(
      "invalid evaluator options",
    );
    await expect(main([catalog, "--offset=-1"])).rejects.toThrow(
      "offset must be",
    );
    await expect(main([catalog, "--split", "train"])).rejects.toThrow(
      "no routing scenarios selected",
    );
    await expect(main([catalog, "extra"])).rejects.toThrow(
      "unrecognized arguments",
    );
  });
});
