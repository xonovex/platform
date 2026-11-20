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
import type {EvaluatorPorts} from "@xonovex/script-moon-skill-eval-common/harness-ports";
import type {ProcessInvocation} from "@xonovex/script-moon-skill-eval-common/process-runner";
import {beforeEach, describe, expect, it} from "vitest";
import {main} from "../../../src/evaluate.js";

const QUERIES = [
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
];

const claudeStream = (triggered: boolean): string =>
  [
    JSON.stringify({type: "system", subtype: "init", skills: ["test-skill"]}),
    ...(triggered
      ? [
          JSON.stringify({
            message: {
              content: [
                {type: "tool_use", name: "Skill", input: {skill: "test-skill"}},
              ],
            },
          }),
        ]
      : []),
  ].join("\n");

const codexStream = (triggered: boolean): string =>
  JSON.stringify({
    type: "item.completed",
    item: {
      type: "agent_message",
      text: triggered ? "XONOVEX_SKILL_TRIGGERED" : "NOT_APPLICABLE",
    },
  });

// A harness that fires the skill on the positive query and nothing else, which is
// the decision the evaluator scores. The query opens the final argument of both
// harnesses, so both recognize it the same way.
const triggeredOn = (invocation: ProcessInvocation): ScriptedProcess => {
  const triggered = (invocation.args.at(-1) ?? "").startsWith("positive");
  return {
    stdout: invocation.executable.endsWith("codex")
      ? codexStream(triggered)
      : claudeStream(triggered),
  };
};

const DIRECTORY = "/skill";

// The skill a run resolves from, holding the SKILL.md that names the skill and
// the queries the evaluator scores, beside the harness binaries the run locates on
// its PATH. Those binaries are never spawned: the process runner answers for them.
const skillTree = (extra: Readonly<Record<string, string>> = {}): FileSystem =>
  memoryFileSystem({
    files: {
      [join(DIRECTORY, "SKILL.md")]:
        "---\nname: test-skill\ndescription: Use for tests.\n---\n",
      [join(DIRECTORY, "eval-queries.json")]: JSON.stringify(QUERIES),
      "/usr/bin/claude": "",
      "/usr/bin/codex": "",
      ...extra,
    },
    executables: ["/usr/bin/claude", "/usr/bin/codex"],
  });

const ports = (
  fs: FileSystem,
  answer: (invocation: ProcessInvocation) => ScriptedProcess = triggeredOn,
  output: MemoryOutput = memoryOutput(),
): EvaluatorPorts => ({
  fs,
  scratch: memoryScratch(fs),
  runProcess: memoryProcessRunner(answer).run,
  environment: {PATH: "/usr/bin"},
  output,
  platform: "linux",
});

describe("main", () => {
  it("prints the usage to stdout for help", async () => {
    const output = memoryOutput();

    await expect(
      main(["--help"], DIRECTORY, ports(skillTree(), triggeredOn, output)),
    ).resolves.toBe(0);
    expect(output.stdoutLines()[0]).toMatch(/^Usage: moon-skill-eval-triggers/);
  });

  it("fails cleanly when the query file is missing", async () => {
    const output = memoryOutput();

    await expect(
      main(
        ["missing-queries.json", "testing"],
        DIRECTORY,
        ports(
          memoryFileSystem({directories: [DIRECTORY]}),
          triggeredOn,
          output,
        ),
      ),
    ).resolves.toBe(2);
    expect(output.stderrLines()).toEqual([
      "Error: queries file not found: /skill/missing-queries.json",
    ]);
  });

  it("reports an unknown flag and an extra positional argument as usage errors", async () => {
    for (const argv of [["--unknown"], ["a", "b", "all", "extra"]]) {
      const output = memoryOutput();

      await expect(
        main(argv, DIRECTORY, ports(skillTree(), triggeredOn, output)),
      ).resolves.toBe(2);
      expect(output.stderrLines()[0]).toMatch(
        /\nmoon-skill-eval-triggers: error: unrecognized arguments: (--unknown|extra)$/,
      );
    }
  });

  it("rejects a harness CLI that does not start", async () => {
    const output = memoryOutput();

    await expect(
      main(
        ["--harness", "codex", QUERIES_FILE, "test-skill"],
        DIRECTORY,
        ports(
          skillTree(),
          () => ({spawnError: "spawn EACCES", code: null}),
          output,
        ),
      ),
    ).resolves.toBe(2);
    expect(output.stderrLines()).toEqual([
      "Error: 'codex' CLI not found in PATH",
    ]);
  });
});

// A Claude run loads the skill from a plugin directory, so every case that uses
// that harness needs one before the evaluator will accept its arguments.
const PLUGIN_DIRECTORY = join(DIRECTORY, "plugin");
const QUERIES_FILE = join(DIRECTORY, "eval-queries.json");

const evaluatorTree = (): FileSystem =>
  skillTree({
    [join(PLUGIN_DIRECTORY, ".claude-plugin", "plugin.json")]: JSON.stringify({
      name: "test-plugin",
      version: "0.0.0",
      dependencies: [],
      skills: ["./skills/test-skill"],
    }),
  });

describe("the trigger evaluator", () => {
  let fs: FileSystem;
  let output: MemoryOutput;

  beforeEach(() => {
    fs = evaluatorTree();
    output = memoryOutput();
  });

  it("runs both trigger decisions and writes evidence", async () => {
    const workspace = join(DIRECTORY, "evidence");

    const exitCode = await main(
      [
        "--runs",
        "1",
        "--workspace",
        workspace,
        "--plugin-dir",
        PLUGIN_DIRECTORY,
        QUERIES_FILE,
        "test-skill",
      ],
      DIRECTORY,
      ports(fs, triggeredOn, output),
    );

    expect(exitCode).toBe(0);
    expect(
      output.stdoutLines().map((line): unknown => JSON.parse(line)),
    ).toMatchObject([
      {query: "positive", pass: true},
      {query: "negative", pass: true},
    ]);
    expect(output.stderrLines().at(-1)).toBe("passed: 2 / 2   failed: 0");
    expect(
      JSON.parse(fs.readText(join(workspace, "summary.json"))),
    ).toMatchObject({
      skill: "test-skill",
      split: "all",
      queries: 2,
      runs_per_query: 1,
      passed: 2,
      failed: 0,
    });
    expect(
      fs.readText(join(workspace, "results.jsonl")).trim().split("\n"),
    ).toHaveLength(2);
  });

  it("runs trigger decisions through the Codex harness", async () => {
    const workspace = join(DIRECTORY, "codex-evidence");
    const exitCode = await main(
      [
        "--harness",
        "codex",
        "--runs",
        "1",
        "--workspace",
        workspace,
        QUERIES_FILE,
        "test-skill",
      ],
      DIRECTORY,
      ports(fs),
    );

    expect(exitCode).toBe(0);
    expect(
      JSON.parse(fs.readText(join(workspace, "summary.json"))),
    ).toMatchObject({
      skill: "test-skill",
      harness: "codex",
      model: "gpt-5.3-codex",
      queries: 2,
      passed: 2,
      failed: 0,
    });
  });

  it("fails the run when a query decides the wrong way", async () => {
    const workspace = join(DIRECTORY, "failing-evidence");

    const exitCode = await main(
      [
        "--runs",
        "1",
        "--workspace",
        workspace,
        "--plugin-dir",
        PLUGIN_DIRECTORY,
        QUERIES_FILE,
        "test-skill",
      ],
      DIRECTORY,
      ports(fs, () => ({stdout: claudeStream(false)})),
    );

    expect(exitCode).toBe(1);
    expect(
      JSON.parse(fs.readText(join(workspace, "summary.json"))),
    ).toMatchObject({passed: 1, failed: 1});
  });

  it("invalidates the run when the harness reports an infrastructure failure", async () => {
    const workspace = join(DIRECTORY, "invalid-evidence");

    const exitCode = await main(
      [
        "--runs",
        "1",
        "--workspace",
        workspace,
        "--plugin-dir",
        PLUGIN_DIRECTORY,
        QUERIES_FILE,
        "test-skill",
      ],
      DIRECTORY,
      ports(fs, () => ({code: 3}), output),
    );

    expect(exitCode).toBe(2);
    expect(output.stderrLines()).toEqual(
      expect.arrayContaining([
        "retrying after transient failure (2/3): claude exited 3",
        'Error: trigger infrastructure failure for query "positive": claude exited 3',
      ]),
    );
    expect(
      JSON.parse(fs.readText(join(workspace, "invalid-run.json"))),
    ).toEqual({query: "positive", error: "claude exited 3"});
  });
});
