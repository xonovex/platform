import {join} from "node:path";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {memoryFileSystem} from "@xonovex/script-moon-common/file-system-memory";
import type {TriggerOutcome} from "@xonovex/script-moon-skill-eval-common/trigger-scan";
import {afterEach, beforeEach, describe, expect, it, vi} from "vitest";
import {main, type TriggerDependencies} from "../../../src/evaluate.js";

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

// A harness that fires the skill on the positive query and nothing else, which is
// the decision the evaluator scores. The query is the last thing a Codex prompt
// carries, so both harnesses recognize it the same way.
const triggeredOn = (text: string): TriggerOutcome =>
  text.startsWith("positive")
    ? {triggered: true, error: null, selected: "target"}
    : {triggered: false, error: null, selected: "none"};

const DIRECTORY = "/skill";

// The guide a run resolves from, holding the SKILL.md that names the skill and
// the queries the evaluator scores.
const guideTree = (extra: Readonly<Record<string, string>> = {}): FileSystem =>
  memoryFileSystem({
    files: {
      [join(DIRECTORY, "SKILL.md")]:
        "---\nname: test-skill\ndescription: Use for tests.\n---\n",
      [join(DIRECTORY, "eval-queries.json")]: JSON.stringify(QUERIES),
      ...extra,
    },
  });

const dependencies = (
  fs: FileSystem,
  overrides: Partial<TriggerDependencies> = {},
): TriggerDependencies => ({
  fs,
  checkTriggered: (query) => Promise.resolve(triggeredOn(query)),
  checkCodexTriggered: (options) => Promise.resolve(triggeredOn(options.query)),
  // The harness binary is never spawned here, so the probe reports it as present
  // and the resolver hands back a path that is only ever passed through.
  configOptions: {
    executableRuns: () => true,
    resolveExecutablePath: (command) => `/usr/bin/${command}`,
    workingDirectory: DIRECTORY,
  },
  discard: () => {
    // Nothing to discard: each case builds its own tree.
  },
  ...overrides,
});

describe("main", () => {
  it("fails cleanly when the query file is missing", async () => {
    await expect(
      main(
        ["missing-queries.json", "testing"],
        dependencies(memoryFileSystem({directories: [DIRECTORY]})),
      ),
    ).resolves.toBe(2);
  });

  it("reports invalid positional arguments as a usage error", async () => {
    vi.spyOn(process.stderr, "write").mockImplementation(() => true);

    await expect(
      main(["a", "b", "all", "extra"], dependencies(guideTree())),
    ).rejects.toThrow('process.exit unexpectedly called with "2"');

    vi.restoreAllMocks();
  });
});

// A Claude run loads the skill from a plugin directory, so every case that uses
// that harness needs one before the evaluator will accept its arguments.
const PLUGIN_DIRECTORY = join(DIRECTORY, "plugin");
const QUERIES_FILE = join(DIRECTORY, "eval-queries.json");

const evaluatorTree = (): FileSystem =>
  guideTree({
    [join(PLUGIN_DIRECTORY, ".claude-plugin", "plugin.json")]: JSON.stringify({
      name: "test-plugin",
      version: "0.0.0",
      dependencies: [],
      skills: ["./skills/test-skill"],
    }),
  });

describe("the trigger evaluator", () => {
  let fs: FileSystem;

  beforeEach(() => {
    fs = evaluatorTree();
    vi.spyOn(process.stderr, "write").mockImplementation(() => true);
    vi.spyOn(process.stdout, "write").mockImplementation(() => true);
  });

  afterEach(() => {
    vi.restoreAllMocks();
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
      dependencies(fs),
    );

    expect(exitCode).toBe(0);
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
      dependencies(fs),
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
      dependencies(fs, {
        checkTriggered: () =>
          Promise.resolve({triggered: false, error: null, selected: "none"}),
      }),
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
      dependencies(fs, {
        checkTriggered: () =>
          Promise.resolve({triggered: false, error: "claude exited 3"}),
      }),
    );

    expect(exitCode).toBe(2);
  });
});
