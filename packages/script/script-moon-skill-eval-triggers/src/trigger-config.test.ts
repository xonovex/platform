import {mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach, beforeEach, describe, expect, it} from "vitest";
import {resolveTriggerConfig, type TriggerConfig} from "./trigger-config.js";

const query = (name: string, split: "train" | "validation" = "train") => ({
  query: name,
  should_trigger: true,
  rationale: "matches",
  split,
});

type TriggerConfigResult = ReturnType<typeof resolveTriggerConfig>;
const defaultEnvironment = {PLUGIN_DIR: "plugin"} as const;

const successfulConfig = (result: TriggerConfigResult): TriggerConfig => {
  expect(result.success).toBe(true);
  if (!result.success) throw new Error(result.error);
  return result.data;
};

const failureMessage = (
  result: TriggerConfigResult,
  kind: "runtime" | "usage",
): string => {
  expect(result).toMatchObject({success: false, kind});
  if (result.success) throw new Error("expected configuration failure");
  return result.error;
};

describe("trigger configuration", () => {
  let directory: string;

  beforeEach(() => {
    directory = mkdtempSync(join(tmpdir(), "skill-trigger-config-"));
    writeFileSync(
      join(directory, "SKILL.md"),
      "---\nname: 'plugin:test-skill'\n---\n",
    );
    writeFileSync(
      join(directory, "eval-queries.json"),
      JSON.stringify([query("train"), query("validation", "validation")]),
    );
    const manifestDirectory = join(directory, "plugin", ".claude-plugin");
    mkdirSync(manifestDirectory, {recursive: true});
    writeFileSync(
      join(manifestDirectory, "plugin.json"),
      JSON.stringify({
        name: "test-plugin",
        version: "0.0.0",
        dependencies: [],
        skills: ["./skills/test-skill"],
      }),
    );
  });

  afterEach(() => {
    rmSync(directory, {recursive: true, force: true});
  });

  const resolveConfig = (
    argv: readonly string[] = [],
    environment: Readonly<
      Record<string, string | undefined>
    > = defaultEnvironment,
  ) =>
    resolveTriggerConfig(argv, {
      environment,
      executableRuns: () => true,
      resolveExecutablePath: () => "/bin/claude",
      workingDirectory: directory,
    });

  it("resolves frontmatter, defaults, queries, and plugin arguments", () => {
    const config = successfulConfig(resolveConfig());

    expect(config).toMatchObject({
      skillName: "plugin:test-skill",
      shortName: "test-skill",
      split: "all",
      runs: 3,
      threshold: 0.5,
      budget: 0.05,
      harness: "claude",
      model: "haiku",
      harnessExecutable: "/bin/claude",
      queryCount: 2,
      maxBatchModelRuns: 6,
      workspace: undefined,
    });
    expect(config.queryBatches).toHaveLength(1);
    expect(config.harnessArgs).toContain(join(directory, "plugin"));
  });

  it("honors explicit CLI values and bounded batches", () => {
    const config = successfulConfig(
      resolveConfig([
        "--runs=2",
        "--threshold",
        "0.75",
        "--model",
        "sonnet",
        "--batch-size",
        "1",
        "--workspace",
        "evidence",
        "eval-queries.json",
        "explicit-skill",
        "validation",
      ]),
    );

    expect(config).toMatchObject({
      skillName: "explicit-skill",
      split: "validation",
      runs: 2,
      threshold: 0.75,
      model: "sonnet",
      queryCount: 1,
      maxBatchModelRuns: 2,
      workspace: join(directory, "evidence"),
    });
    expect(config.queryBatches).toHaveLength(1);
  });

  it("selects the Codex harness with isolated non-interactive arguments", () => {
    const config = successfulConfig(
      resolveConfig(["--harness", "codex", "--model", "gpt-test"]),
    );

    expect(config).toMatchObject({
      harness: "codex",
      harnessExecutable: "/bin/claude",
      model: "gpt-test",
    });
    expect(config.harnessArgs).toEqual(
      expect.arrayContaining([
        "exec",
        "--json",
        "--ephemeral",
        "--sandbox",
        "read-only",
        "--ignore-user-config",
        "--ignore-rules",
      ]),
    );
  });

  it("classifies invalid positionals and options as usage errors", () => {
    expect(
      failureMessage(resolveConfig(["a", "b", "all", "extra"]), "usage"),
    ).toContain("unrecognized arguments");
    expect(
      failureMessage(
        resolveConfig([
          "--split",
          "train",
          "eval-queries.json",
          "skill",
          "validation",
        ]),
        "usage",
      ),
    ).toContain("split must not be provided twice");
    expect(
      failureMessage(
        resolveConfig(["eval-queries.json", "skill", "unknown"]),
        "usage",
      ),
    ).toContain("invalid choice");
    expect(failureMessage(resolveConfig(["--runs", "4"]), "usage")).toContain(
      "invalid evaluator options",
    );

    rmSync(join(directory, "SKILL.md"));
    expect(failureMessage(resolveConfig(), "usage")).toContain(
      "arguments are required: skill_name",
    );
  });

  it("rejects missing files, executables, and plugin directories", () => {
    expect(
      failureMessage(resolveConfig(["missing.json", "skill"]), "runtime"),
    ).toContain("queries file not found");
    expect(
      failureMessage(
        resolveTriggerConfig([], {
          environment: {PLUGIN_DIR: "plugin"},
          executableRuns: () => true,
          resolveExecutablePath: () => void 0,
          workingDirectory: directory,
        }),
        "runtime",
      ),
    ).toContain("CLI not found");
    expect(
      failureMessage(
        resolveTriggerConfig([], {
          environment: {PLUGIN_DIR: "plugin"},
          executableRuns: () => false,
          resolveExecutablePath: () => "/bin/claude",
          workingDirectory: directory,
        }),
        "runtime",
      ),
    ).toContain("CLI not found");
    expect(
      failureMessage(resolveConfig([], {PLUGIN_DIR: "missing"}), "runtime"),
    ).toContain("target plugin directory is invalid");
  });

  it("rejects malformed queries and empty split selections", () => {
    const queriesFile = join(directory, "eval-queries.json");
    writeFileSync(queriesFile, "not-json");
    expect(failureMessage(resolveConfig(), "runtime")).toContain(
      "invalid JSON",
    );

    writeFileSync(queriesFile, JSON.stringify([]));
    expect(failureMessage(resolveConfig(), "runtime")).toContain(
      "invalid queries",
    );

    writeFileSync(queriesFile, JSON.stringify([query("train")]));
    expect(
      failureMessage(
        resolveConfig(["eval-queries.json", "skill", "validation"]),
        "runtime",
      ),
    ).toContain("has no queries");
  });

  it("rejects oversized batches and accepts a bounded equivalent", () => {
    writeFileSync(
      join(directory, "eval-queries.json"),
      JSON.stringify(
        Array.from({length: 9}, (_, index) => query(String(index))),
      ),
    );

    expect(failureMessage(resolveConfig(), "runtime")).toContain(
      "would launch 27 model runs",
    );
    expect(resolveConfig(["--batch-size", "8"])).toMatchObject({
      success: true,
    });
  });
});
