import {join} from "node:path";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {memoryFileSystem} from "@xonovex/script-moon-common/file-system-memory";
import {beforeEach, describe, expect, it} from "vitest";
import {
  resolveTriggerConfig,
  type TriggerConfig,
} from "../../../src/trigger-config.js";

const query = (name: string, split: "train" | "validation" = "train") => ({
  query: name,
  should_trigger: true,
  rationale: "matches",
  split,
});

const negative = (name: string) => ({
  query: name,
  should_trigger: false,
  rationale: "near miss",
  split: "train" as const,
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

const DIRECTORY = "/skill";

// The guide a run resolves from: its SKILL.md, its queries, and the plugin
// manifest a Claude run loads the skill through.
const guideTree = ({skill = true}: {skill?: boolean} = {}): FileSystem =>
  memoryFileSystem({
    files: {
      ...(skill
        ? {
            [join(DIRECTORY, "SKILL.md")]:
              "---\nname: 'plugin:test-skill'\n---\n",
          }
        : {}),
      [join(DIRECTORY, "eval-queries.json")]: JSON.stringify([
        query("train"),
        query("validation", "validation"),
      ]),
      [join(DIRECTORY, "plugin", ".claude-plugin", "plugin.json")]:
        JSON.stringify({
          name: "test-plugin",
          version: "0.0.0",
          dependencies: [],
          skills: ["./skills/test-skill"],
        }),
    },
  });

describe("trigger configuration", () => {
  const directory = DIRECTORY;
  let fs: FileSystem;

  beforeEach(() => {
    fs = guideTree();
  });

  const resolveConfig = (
    argv: readonly string[] = [],
    environment: Readonly<
      Record<string, string | undefined>
    > = defaultEnvironment,
  ) =>
    resolveTriggerConfig(argv, {
      fs,
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
      model: "claude-haiku-4-5-20251001",
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

  it("uses the pinned Codex model when no override is provided", () => {
    const config = successfulConfig(resolveConfig(["--harness", "codex"]));

    expect(config.model).toBe("gpt-5.3-codex");
    expect(config.harnessArgs).toContain("gpt-5.3-codex");
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

    fs = guideTree({skill: false});
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
          fs,
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
          fs,
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
    fs.writeFile(queriesFile, "not-json");
    expect(failureMessage(resolveConfig(), "runtime")).toContain(
      "invalid JSON",
    );

    fs.writeFile(queriesFile, JSON.stringify([]));
    expect(failureMessage(resolveConfig(), "runtime")).toContain(
      "invalid queries",
    );

    fs.writeFile(queriesFile, JSON.stringify([query("train")]));
    expect(
      failureMessage(
        resolveConfig(["eval-queries.json", "skill", "validation"]),
        "runtime",
      ),
    ).toContain("has no queries");
  });

  const writeCatalog = (
    entries: readonly {
      readonly directory: string;
      readonly name: string;
      readonly queries: readonly unknown[];
    }[],
  ): string => {
    const catalogRoot = join(directory, "catalog");
    for (const entry of entries) {
      const guideDirectory = join(catalogRoot, entry.directory, entry.name);
      fs.makeDirectory(guideDirectory);
      fs.writeFile(
        join(guideDirectory, "SKILL.md"),
        `---\nname: ${entry.name}\n---\n`,
      );
      fs.writeFile(
        join(guideDirectory, "eval-queries.json"),
        JSON.stringify(entry.queries),
      );
    }
    return catalogRoot;
  };

  it("leaves a query another catalog skill owns to the routing evaluator", () => {
    fs.writeFile(
      join(directory, "eval-queries.json"),
      JSON.stringify([query("own concept"), negative("owned elsewhere")]),
    );
    const catalogRoot = writeCatalog([
      {
        directory: "skill-test",
        name: "test-skill",
        queries: [query("own concept"), negative("owned elsewhere")],
      },
      {
        directory: "skill-other",
        name: "other-skill",
        queries: [query("owned elsewhere")],
      },
    ]);

    const config = successfulConfig(
      resolveConfig(["--catalog-root", catalogRoot]),
    );

    expect(config).toMatchObject({queryCount: 1, deferredToRouting: 1});
    expect(config.queryBatches.flat().map((entry) => entry.query)).toEqual([
      "own concept",
    ]);
  });

  it("keeps a negative no catalog skill claims", () => {
    fs.writeFile(
      join(directory, "eval-queries.json"),
      JSON.stringify([query("own concept"), negative("unclaimed")]),
    );
    const catalogRoot = writeCatalog([
      {
        directory: "skill-test",
        name: "test-skill",
        queries: [query("own concept"), negative("unclaimed")],
      },
    ]);

    const config = successfulConfig(
      resolveConfig(["--catalog-root", catalogRoot]),
    );

    expect(config).toMatchObject({queryCount: 2, deferredToRouting: 0});
  });

  it("scores every query when no catalog root is given", () => {
    const config = successfulConfig(resolveConfig());

    expect(config).toMatchObject({deferredToRouting: 0});
  });

  it("reports an unreadable catalog root", () => {
    expect(
      failureMessage(
        resolveConfig(["--catalog-root", join(directory, "absent")]),
        "runtime",
      ),
    ).toContain("catalog root unreadable");
  });

  it("rejects oversized batches and accepts a bounded equivalent", () => {
    fs.writeFile(
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
