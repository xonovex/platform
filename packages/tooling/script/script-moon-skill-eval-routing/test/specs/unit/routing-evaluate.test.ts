import {join} from "node:path";
import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import {
  memoryOutput,
  type MemoryOutput,
} from "@xonovex/script-moon-common/adapters/output-memory";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {
  memoryProcessRunner,
  type ScriptedProcess,
} from "@xonovex/script-moon-skill-eval-common/adapters/process-runner-memory";
import {memoryScratch} from "@xonovex/script-moon-skill-eval-common/adapters/scratch-memory";
import type {Environment} from "@xonovex/script-moon-skill-eval-common/environment";
import type {EvaluatorPorts} from "@xonovex/script-moon-skill-eval-common/harness-ports";
import type {ProcessInvocation} from "@xonovex/script-moon-skill-eval-common/process-runner";
import {beforeEach, describe, expect, it} from "vitest";
import {main} from "../../../src/routing-evaluate.js";

const CATALOG = "/catalog";

// A catalog of two competing skills: alpha owns the shared query, beta carries it
// as a negative, which is the smallest pairing a routing scenario needs. The tree
// also holds the harness binaries a run locates on its PATH; the process runner
// answers for them, so they are never spawned.
const catalogTree = (): FileSystem => {
  const files: Record<string, string> = {
    "/usr/bin/claude": "",
    "/usr/bin/codex": "",
  };
  for (const [name, shouldTrigger] of [
    ["alpha", true],
    ["beta", false],
  ] as const) {
    const plugin = join(CATALOG, `plugin-${name}`);
    const skill = join(plugin, "skills", `using-${name}`);
    files[join(plugin, ".claude-plugin", "plugin.json")] = JSON.stringify({
      name: `xonovex-${name}`,
      version: "0.0.0",
      skills: "./skills/",
    });
    files[join(skill, "SKILL.md")] =
      `---\nname: using-${name}\ndescription: Use for ${name}.\n---\n`;
    files[join(skill, "eval-queries.json")] = JSON.stringify([
      {
        query: "shared request",
        should_trigger: shouldTrigger,
        split: "validation",
      },
    ]);
  }
  return memoryFileSystem({
    files,
    executables: ["/usr/bin/claude", "/usr/bin/codex"],
  });
};

const line = (event: unknown): string => JSON.stringify(event);

const init = (...skills: readonly string[]): string =>
  line({type: "system", subtype: "init", skills});

const invoke = (skill: string): string =>
  line({
    message: {content: [{type: "tool_use", name: "Skill", input: {skill}}]},
  });

const codexMessage = (text: string): string =>
  line({type: "item.completed", item: {type: "agent_message", text}});

type Harness = "claude" | "codex";

// How each harness reports a run: the owner chosen, a competitor chosen, or an
// owner the harness could not load, which is an infrastructure failure.
const SELECTED: Readonly<Record<Harness, string>> = {
  claude: [init("using-alpha", "using-beta"), invoke("using-alpha")].join("\n"),
  codex: codexMessage("XONOVEX_SKILL_TRIGGERED"),
};

const NOT_SELECTED: Readonly<Record<Harness, string>> = {
  claude: [
    init("using-alpha", "using-beta"),
    invoke("using-beta"),
    line({message: {content: [{type: "tool_result", content: "loaded"}]}}),
  ].join("\n"),
  codex: codexMessage("NOT_APPLICABLE"),
};

const UNAVAILABLE: Readonly<Record<Harness, string>> = {
  claude: init(),
  codex: "",
};

const answering =
  (stdout: Readonly<Record<Harness, string>>) =>
  (invocation: ProcessInvocation): ScriptedProcess => ({
    stdout:
      stdout[invocation.executable.endsWith("codex") ? "codex" : "claude"],
  });

// Evidence is written into the same tree the catalog is read from, so a case
// asserts on what a run produced without a disk.
const ports = (
  fs: FileSystem,
  stdout: Readonly<Record<Harness, string>> = SELECTED,
  output: MemoryOutput = memoryOutput(),
  environment: Environment = {},
): EvaluatorPorts => ({
  fs,
  scratch: memoryScratch(fs),
  runProcess: memoryProcessRunner(answering(stdout)).run,
  environment: {PATH: "/usr/bin", ...environment},
  output,
  platform: "linux",
});

describe("catalog routing evaluator", () => {
  const catalog = CATALOG;
  let fs: FileSystem;

  const readWorkspace = (workspace: string, file: string): string =>
    fs.readText(join(workspace, file));

  // The error line a run that cannot start reports, with its exit code.
  const startFailure = async (
    argv: readonly string[],
    environment: Environment = {},
  ): Promise<{exitCode: number; stderr: readonly string[]}> => {
    const output = memoryOutput();
    const exitCode = await main(
      argv,
      "/",
      ports(fs, SELECTED, output, environment),
    );
    return {exitCode, stderr: output.stderrLines()};
  };

  beforeEach(() => {
    fs = catalogTree();
  });

  it("loads competing Claude plugins and writes passing evidence", async () => {
    const workspace = join("/evidence", "claude-evidence");

    const output = memoryOutput();

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
      "/",
      ports(fs, SELECTED, output),
    );

    expect(exitCode).toBe(0);
    expect(output.stderrLines()).toEqual(["batch 1/1: 1 routing scenarios"]);
    expect(
      output.stdoutLines().map((record): unknown => JSON.parse(record)),
    ).toEqual([
      {
        query: "shared request",
        split: "validation",
        expected_skill: "using-alpha",
        candidate_skills: ["using-alpha", "using-beta"],
        owner_selections: 1,
        runs: 1,
        selection_rate: 1,
        pass: true,
      },
    ]);
    expect(JSON.parse(readWorkspace(workspace, "summary.json"))).toMatchObject({
      harness: "claude",
      model: "claude-haiku-5-5",
      catalog_scenarios: 1,
      selected_scenarios: 1,
      passed: 1,
      failed: 0,
    });
    expect(readWorkspace(workspace, "results.jsonl")).toContain(
      '"candidate_skills":["using-alpha","using-beta"]',
    );
  });

  it("stages competing Codex skills and honors limit and offset", async () => {
    const workspace = join("/evidence", "codex-evidence");

    const exitCode = await main(
      [
        catalog,
        "--harness=codex",
        "--workspace",
        workspace,
        "--offset=0",
        "--limit=1",
        "--model=gpt-test",
      ],
      "/",
      ports(fs),
    );

    expect(exitCode).toBe(0);
    expect(JSON.parse(readWorkspace(workspace, "summary.json"))).toMatchObject({
      harness: "codex",
      model: "gpt-test",
      selected_scenarios: 1,
    });
  });

  it("filters routing scenarios to selected expected owners", async () => {
    const workspace = join("/evidence", "owner-evidence");

    const exitCode = await main(
      [catalog, "--workspace", workspace, "--owners", "using-alpha"],
      "/",
      ports(fs),
    );

    expect(exitCode).toBe(0);
    expect(JSON.parse(readWorkspace(workspace, "summary.json"))).toMatchObject({
      selected_owners: ["using-alpha"],
      selected_scenarios: 1,
    });
    await expect(
      startFailure([catalog, "--owners", "using-beta"]),
    ).resolves.toEqual({
      exitCode: 2,
      stderr: ["Error: no routing scenarios selected"],
    });
  });

  it("returns one when a healthy run selects no owner", async () => {
    await expect(
      main(
        [catalog, "--workspace", join("/evidence", "failed-evidence")],
        "/",
        ports(fs, NOT_SELECTED),
      ),
    ).resolves.toBe(1);
  });

  it("returns two and records an infrastructure failure", async () => {
    const workspace = join("/evidence", "invalid-evidence");

    await expect(
      main([catalog, "--workspace", workspace], "/", ports(fs, UNAVAILABLE)),
    ).resolves.toBe(2);
    expect(readWorkspace(workspace, "invalid-run.json")).toContain(
      "target skill unavailable",
    );
  });

  it("prints the help to stdout", async () => {
    const output = memoryOutput();

    await expect(
      main(["--help"], "/", ports(fs, SELECTED, output)),
    ).resolves.toBe(0);
    expect(output.stdoutLines()[0]).toMatch(/^Usage: moon-skill-eval-routing/);
  });

  it("reads the harness and its model from the environment", async () => {
    const workspace = join("/evidence", "environment-evidence");

    const exitCode = await main(
      [catalog, "--workspace", workspace],
      "/",
      ports(fs, SELECTED, memoryOutput(), {
        SKILL_EVAL_HARNESS: "codex",
        CODEX_MODEL: "gpt-env",
      }),
    );

    expect(exitCode).toBe(0);
    expect(JSON.parse(readWorkspace(workspace, "summary.json"))).toMatchObject({
      harness: "codex",
      model: "gpt-env",
    });
  });

  it("resolves the catalog and the workspace from the working directory", async () => {
    const exitCode = await main(["catalog"], "/", ports(fs));

    expect(exitCode).toBe(0);
    expect(fs.isFile("/.skill-eval-results/routing/claude/summary.json")).toBe(
      true,
    );
  });

  it("rejects invalid selection and evaluator options", async () => {
    for (const [argv, error] of [
      [
        ["--harness", "other"],
        "invalid harness 'other'; expected claude or codex",
      ],
      [["--split", "other"], "invalid split"],
      [["--runs", "4"], "invalid evaluator options"],
      [["--offset=-1"], "offset must be a non-negative integer"],
      [["--limit=x"], "limit must be a non-negative integer"],
      [["--split", "train"], "no routing scenarios selected"],
      [["extra"], "unrecognized arguments: extra"],
    ] as const) {
      const {exitCode, stderr} = await startFailure([catalog, ...argv]);

      expect(exitCode).toBe(2);
      expect(stderr[0]).toContain(`Error: ${error}`);
    }
  });

  it("rejects a batch that would launch more model runs than the cap", async () => {
    // Nine shared queries at three runs each exceed the cap in one batch.
    for (const [name, shouldTrigger] of [
      ["alpha", true],
      ["beta", false],
    ] as const) {
      fs.writeFile(
        join(
          CATALOG,
          `plugin-${name}`,
          "skills",
          `using-${name}`,
          "eval-queries.json",
        ),
        JSON.stringify(
          Array.from({length: 9}, (_, index) => ({
            query: `shared request ${String(index)}`,
            should_trigger: shouldTrigger,
            split: "validation",
          })),
        ),
      );
    }

    await expect(
      startFailure([catalog, "--runs", "3", "--batch-size", "9"]),
    ).resolves.toEqual({
      exitCode: 2,
      stderr: [
        "Error: routing eval batch would launch 27 model runs; maximum is 24",
      ],
    });
  });

  it("rejects an unreadable catalog and a harness missing from PATH", async () => {
    await expect(startFailure(["/missing"])).resolves.toMatchObject({
      exitCode: 2,
    });
    await expect(
      startFailure([catalog, "--harness", "codex"], {PATH: "/opt/bin"}),
    ).resolves.toEqual({
      exitCode: 2,
      stderr: ["Error: 'codex' CLI not found in PATH"],
    });
  });
});
