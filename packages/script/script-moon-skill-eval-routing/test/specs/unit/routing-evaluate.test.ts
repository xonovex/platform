import {join} from "node:path";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {memoryFileSystem} from "@xonovex/script-moon-common/file-system-memory";
import type {TriggerOutcome} from "@xonovex/script-moon-skill-eval-common/trigger-scan";
import {afterEach, beforeEach, describe, expect, it, vi} from "vitest";
import {main, type RoutingDependencies} from "../../../src/routing-evaluate.js";

const CATALOG = "/catalog";

// A catalog of two competing skills: alpha owns the shared query, beta carries it
// as a negative, which is the smallest pairing a routing scenario needs.
const catalogTree = (): FileSystem => {
  const files: Record<string, string> = {};
  for (const [name, shouldTrigger] of [
    ["alpha", true],
    ["beta", false],
  ] as const) {
    const plugin = join(CATALOG, `skill-${name}`);
    const guide = join(plugin, `${name}-guide`);
    files[join(plugin, ".claude-plugin", "plugin.json")] = JSON.stringify({
      name: `xonovex-skill-${name}`,
      version: "0.0.0",
      skills: [`./${name}-guide`],
    });
    files[join(guide, "SKILL.md")] =
      `---\nname: ${name}-guide\ndescription: Use for ${name}.\n---\n`;
    files[join(guide, "eval-queries.json")] = JSON.stringify([
      {
        query: "shared request",
        should_trigger: shouldTrigger,
        split: "validation",
      },
    ]);
  }
  return memoryFileSystem({files});
};

const SELECTED: TriggerOutcome = {
  triggered: true,
  error: null,
  selected: "target",
};

const NOT_SELECTED: TriggerOutcome = {
  triggered: false,
  error: null,
  selected: "competitor:beta-guide",
};

// Evidence is written into the same tree the catalog is read from, so a case
// asserts on what a run produced without a disk.
const dependencies = (
  fs: FileSystem,
  outcome: TriggerOutcome | (() => TriggerOutcome) = SELECTED,
): RoutingDependencies => {
  const decide = typeof outcome === "function" ? outcome : () => outcome;
  return {
    fs,
    checkTriggered: () => Promise.resolve(decide()),
    checkCodexTriggered: () => Promise.resolve(decide()),
    // The harness binary is never spawned here, so resolution only has to hand
    // back a path the evaluator passes through.
    resolveExecutable: (command) => `/usr/bin/${command}`,
    discard: () => {
      // Nothing to discard: each case builds its own tree.
    },
  };
};

describe("catalog routing evaluator", () => {
  const originalHarness = process.env.SKILL_EVAL_HARNESS;
  const catalog = CATALOG;
  let fs: FileSystem;

  const readWorkspace = (workspace: string, file: string): string =>
    fs.readText(join(workspace, file));

  beforeEach(() => {
    fs = catalogTree();
    delete process.env.SKILL_EVAL_HARNESS;
    vi.spyOn(process.stderr, "write").mockImplementation(() => true);
    vi.spyOn(process.stdout, "write").mockImplementation(() => true);
  });

  afterEach(() => {
    process.env.SKILL_EVAL_HARNESS = originalHarness;
    vi.restoreAllMocks();
  });

  it("loads competing Claude plugins and writes passing evidence", async () => {
    const workspace = join("/evidence", "claude-evidence");

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
      dependencies(fs),
    );

    expect(exitCode).toBe(0);
    expect(JSON.parse(readWorkspace(workspace, "summary.json"))).toMatchObject({
      harness: "claude",
      model: "claude-haiku-4-5-20251001",
      catalog_scenarios: 1,
      selected_scenarios: 1,
      passed: 1,
      failed: 0,
    });
    expect(readWorkspace(workspace, "results.jsonl")).toContain(
      '"candidate_skills":["alpha-guide","beta-guide"]',
    );
  });

  it("stages competing Codex guides and honors limit and offset", async () => {
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
      dependencies(fs),
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
      [catalog, "--workspace", workspace, "--owners", "alpha-guide"],
      dependencies(fs),
    );

    expect(exitCode).toBe(0);
    expect(JSON.parse(readWorkspace(workspace, "summary.json"))).toMatchObject({
      selected_owners: ["alpha-guide"],
      selected_scenarios: 1,
    });
    await expect(
      main([catalog, "--owners", "beta-guide"], dependencies(fs)),
    ).rejects.toThrow("no routing scenarios selected");
  });

  it("returns one when a healthy run selects no owner", async () => {
    await expect(
      main(
        [catalog, "--workspace", join("/evidence", "failed-evidence")],
        dependencies(fs, NOT_SELECTED),
      ),
    ).resolves.toBe(1);
  });

  it("returns two and records an infrastructure failure", async () => {
    const workspace = join("/evidence", "invalid-evidence");

    await expect(
      main(
        [catalog, "--workspace", workspace],
        dependencies(fs, {triggered: false, error: "target skill unavailable"}),
      ),
    ).resolves.toBe(2);
    expect(readWorkspace(workspace, "invalid-run.json")).toContain(
      "target skill unavailable",
    );
  });

  it("rejects invalid selection and evaluator options", async () => {
    const deps = dependencies(fs);
    await expect(main([catalog, "--harness", "other"], deps)).rejects.toThrow(
      "invalid harness",
    );
    await expect(main([catalog, "--split", "other"], deps)).rejects.toThrow(
      "invalid split",
    );
    await expect(main([catalog, "--runs", "4"], deps)).rejects.toThrow(
      "invalid evaluator options",
    );
    await expect(main([catalog, "--offset=-1"], deps)).rejects.toThrow(
      "offset must be",
    );
    await expect(main([catalog, "--split", "train"], deps)).rejects.toThrow(
      "no routing scenarios selected",
    );
    await expect(main([catalog, "extra"], deps)).rejects.toThrow(
      "unrecognized arguments",
    );
  });
});
