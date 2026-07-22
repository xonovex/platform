import {mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {dirname, join} from "node:path";
import {afterEach, describe, expect, it} from "vitest";
import {
  checkCrossPackageLinks,
  checkMarkdownFilesForCrossPackageLinks,
  checkNamedSkillHandoffs,
  checkSkillDependencies,
} from "./cross-package-links.js";
import {type LinkReport} from "./reference-file-links.js";

const created: string[] = [];

const makeSink = (): LinkReport & {fails: string[]; passes: string[]} => {
  const fails: string[] = [];
  const passes: string[] = [];
  return {
    fails,
    passes,
    addFail: (message) => {
      fails.push(message);
    },
    addPass: (message) => {
      passes.push(message);
    },
  };
};

// Build a throwaway repo whose files are written at the given repo-relative
// paths, creating parent directories as needed. Returns the repo root.
const makeRepo = (files: Record<string, string>): string => {
  const root = mkdtempSync(join(tmpdir(), "xpkg-links-"));
  created.push(root);
  for (const [rel, content] of Object.entries(files)) {
    const abs = join(root, rel);
    mkdirSync(dirname(abs), {recursive: true});
    writeFileSync(abs, content);
  }
  return root;
};

const manifest = (
  name: string,
  dependencies: readonly string[] = [],
  kind: "claude" | "codex" = "codex",
  skill = `./${name.replace(/^xonovex-skill-/, "")}-guide`,
): string =>
  JSON.stringify({
    name,
    dependencies,
    skills: kind === "claude" ? [skill] : skill,
  });

const workflowCommandNames = [
  "abandon",
  "create",
  "decide",
  "execute",
  "publish",
  "review",
  "revise",
  "validate",
  "workspace-abandon",
  "workspace-cleanup",
  "workspace-create",
  "workspace-merge",
] as const;

const workflowCommands = (
  overrides: Readonly<Record<string, string>> = {},
): Record<string, string> =>
  Object.fromEntries(
    workflowCommandNames.map((name) => [
      `packages/command/command-workflow/commands/${name}.md`,
      overrides[name] ?? `# ${name}\n`,
    ]),
  );

afterEach(() => {
  while (created.length > 0) {
    const dir = created.pop();
    if (dir !== undefined) rmSync(dir, {recursive: true, force: true});
  }
});

describe("checkMarkdownFilesForCrossPackageLinks", () => {
  it("resolves a boundary-crossing link to an existing target", () => {
    const repo = makeRepo({
      "packages/skill/skill-x/x-guide/references/actors.md": "# Actors\n",
      "packages/command/command-workflow/commands/foo.md":
        "See [actors](../../../skill/skill-x/x-guide/references/actors.md).",
    });
    const source = join(
      repo,
      "packages/command/command-workflow/commands/foo.md",
    );
    const report = makeSink();
    const counts = checkMarkdownFilesForCrossPackageLinks(
      [source],
      repo,
      report,
    );
    expect(counts).toEqual({resolved: 1, broken: 0});
    expect(report.fails).toEqual([]);
  });

  it("fails a boundary-crossing link whose target is missing, naming the source and link", () => {
    const repo = makeRepo({
      "packages/command/command-workflow/commands/foo.md":
        "See [actors](../../../skill/skill-x/x-guide/references/actors.md).",
    });
    const source = join(
      repo,
      "packages/command/command-workflow/commands/foo.md",
    );
    const report = makeSink();
    const counts = checkMarkdownFilesForCrossPackageLinks(
      [source],
      repo,
      report,
    );
    expect(counts.broken).toBe(1);
    expect(report.fails).toHaveLength(1);
    expect(report.fails[0]).toContain(
      "packages/command/command-workflow/commands/foo.md",
    );
    expect(report.fails[0]).toContain(
      "../../../skill/skill-x/x-guide/references/actors.md",
    );
  });

  it("ignores an intra-package link, even a broken one", () => {
    const repo = makeRepo({
      "packages/command/command-workflow/docs/a.md": "See [b](./missing.md).",
    });
    const source = join(repo, "packages/command/command-workflow/docs/a.md");
    const report = makeSink();
    const counts = checkMarkdownFilesForCrossPackageLinks(
      [source],
      repo,
      report,
    );
    expect(counts).toEqual({resolved: 0, broken: 0});
    expect(report.fails).toEqual([]);
  });

  it("skips external, placeholder, ellipsis, and anchor forms", () => {
    const repo = makeRepo({
      "packages/command/command-workflow/docs/skips.md": [
        "[a](https://example.com)",
        "[b](mailto:x@y.z)",
        "[c](../../../skill/<topic>.md)",
        "[d](../../../skill/{topic}.md)",
        "[e](…/pull/PR)",
        "[f](#anchor)",
      ].join("\n"),
    });
    const source = join(
      repo,
      "packages/command/command-workflow/docs/skips.md",
    );
    const report = makeSink();
    const counts = checkMarkdownFilesForCrossPackageLinks(
      [source],
      repo,
      report,
    );
    expect(counts).toEqual({resolved: 0, broken: 0});
    expect(report.fails).toEqual([]);
  });

  it("resolves a boundary-crossing link that carries an in-page fragment", () => {
    const repo = makeRepo({
      "packages/skill/skill-y/y-guide/references/g.md": "# G\n",
      "packages/command/command-workflow/docs/a.md":
        "See [g](../../../skill/skill-y/y-guide/references/g.md#section).",
    });
    const source = join(repo, "packages/command/command-workflow/docs/a.md");
    const report = makeSink();
    const counts = checkMarkdownFilesForCrossPackageLinks(
      [source],
      repo,
      report,
    );
    expect(counts).toEqual({resolved: 1, broken: 0});
    expect(report.fails).toEqual([]);
  });
});

describe("checkCrossPackageLinks", () => {
  it("scans the twelve-command surface and retained planning references", () => {
    const repo = makeRepo({
      ...workflowCommands(),
      "packages/command/command-workflow/README.md":
        "See the [operation model](../../diagram/diagram-agent-workflow/operation-model.png).\n",
      "packages/command/command-workflow/docs/invocation.md":
        "See the [submission boundary](../../../agent/agent-operator-go/README.md).\n",
      "packages/diagram/diagram-agent-workflow/operation-model.png": "image",
      "packages/agent/agent-operator-go/README.md": "# Operator\n",
      "packages/skill/skill-plan/plan-guide/SKILL.md":
        "# Plan\nSee [create](references/plan-create.md).\n",
      "packages/skill/skill-plan/plan-guide/references/plan-create.md":
        "# plan-create\n",
      "packages/skill/skill-plan/.claude-plugin/plugin.json": manifest(
        "xonovex-skill-plan",
        [],
        "claude",
      ),
      "packages/skill/skill-plan/.codex-plugin/plugin.json":
        manifest("xonovex-skill-plan"),
    });
    const report = makeSink();
    checkCrossPackageLinks(repo, report);
    expect(report.fails).toEqual([]);
    expect(report.passes).toContain("cross-package links: 2/2 link(s) resolve");
  });

  it("fails when a discovered SKILL.md points at a moved contract", () => {
    const repo = makeRepo({
      "packages/skill/skill-x/x-guide/SKILL.md":
        "# X\nSee [g](../../skill-y/y-guide/references/g.md).\n",
      "packages/skill/skill-x/.claude-plugin/plugin.json": manifest(
        "xonovex-skill-x",
        [],
        "claude",
      ),
      "packages/skill/skill-x/.codex-plugin/plugin.json":
        manifest("xonovex-skill-x"),
    });
    const report = makeSink();
    checkCrossPackageLinks(repo, report);
    expect(report.passes).not.toContain(
      "cross-package links: 1/1 link(s) resolve",
    );
    expect(report.fails).toHaveLength(1);
    expect(report.fails[0]).toContain(
      "packages/skill/skill-x/x-guide/SKILL.md",
    );
  });
});

describe("checkNamedSkillHandoffs", () => {
  it("resolves catalog guide names and rejects missing handoffs", () => {
    const repo = makeRepo({
      "packages/skill/skill-x/x-guide/SKILL.md":
        "Use **known-guide** and **missing-guide**.\n",
    });
    const source = join(repo, "packages/skill/skill-x/x-guide/SKILL.md");
    const report = makeSink();

    const counts = checkNamedSkillHandoffs(
      [source],
      new Set(["known-guide"]),
      repo,
      report,
    );

    expect(counts).toEqual({resolved: 1, broken: 1});
    expect(report.fails).toContain(
      "skill handoffs: packages/skill/skill-x/x-guide/SKILL.md names missing **missing-guide**",
    );
  });
});

describe("checkSkillDependencies", () => {
  it("accepts matching acyclic manifest pairs", () => {
    const repo = makeRepo({
      "packages/skill/skill-base/base-guide/SKILL.md": "# Base\n",
      "packages/skill/skill-base/.claude-plugin/plugin.json": manifest(
        "xonovex-skill-base",
        [],
        "claude",
      ),
      "packages/skill/skill-base/.codex-plugin/plugin.json":
        manifest("xonovex-skill-base"),
      "packages/skill/skill-child/child-guide/SKILL.md":
        "# Child\nUse **base-guide**.\n",
      "packages/skill/skill-child/.claude-plugin/plugin.json": manifest(
        "xonovex-skill-child",
        ["xonovex-skill-base"],
        "claude",
      ),
      "packages/skill/skill-child/.codex-plugin/plugin.json": manifest(
        "xonovex-skill-child",
        ["xonovex-skill-base"],
      ),
    });
    const report = makeSink();

    checkSkillDependencies(repo, report);

    expect(report.fails).toEqual([]);
    expect(report.passes).toContain(
      "skill dependencies: 2 manifest pair(s) agree with no dangling dependencies or cycles",
    );
  });

  it("rejects mismatched, dangling, and cyclic dependencies", () => {
    const repo = makeRepo({
      "packages/skill/skill-a/.claude-plugin/plugin.json": manifest(
        "xonovex-skill-a",
        ["xonovex-skill-b", "xonovex-skill-missing", "xonovex-skill-other"],
        "claude",
      ),
      "packages/skill/skill-a/.codex-plugin/plugin.json": manifest(
        "xonovex-skill-a",
        ["xonovex-skill-b", "xonovex-skill-missing"],
      ),
      "packages/skill/skill-b/.claude-plugin/plugin.json": manifest(
        "xonovex-skill-b",
        ["xonovex-skill-a"],
        "claude",
      ),
      "packages/skill/skill-b/.codex-plugin/plugin.json": manifest(
        "xonovex-skill-b",
        ["xonovex-skill-a"],
      ),
    });
    const report = makeSink();

    checkSkillDependencies(repo, report);

    expect(report.fails).toContain(
      "skill dependencies: manifest dependencies differ for xonovex-skill-a",
    );
    expect(report.fails).toContain(
      "skill dependencies: xonovex-skill-a depends on missing xonovex-skill-missing",
    );
    expect(report.fails).toContain(
      "skill dependencies: dependency cycle xonovex-skill-a → xonovex-skill-b → xonovex-skill-a",
    );
  });

  it("rejects a discovered skill that has neither manifest", () => {
    const repo = makeRepo({
      "packages/skill/skill-unpackaged/unpackaged-guide/SKILL.md":
        "# Unpackaged\n",
    });
    const report = makeSink();

    checkSkillDependencies(repo, report);

    expect(report.fails).toContain(
      "skill dependencies: packages/skill/skill-unpackaged needs both Claude and Codex manifests",
    );
  });

  it("requires package-derived names and visible hard-dependency handoffs", () => {
    const repo = makeRepo({
      "packages/skill/skill-base/base-guide/SKILL.md": "# Base\n",
      "packages/skill/skill-base/.claude-plugin/plugin.json": manifest(
        "xonovex-skill-wrong",
        [],
        "claude",
      ),
      "packages/skill/skill-base/.codex-plugin/plugin.json": manifest(
        "xonovex-skill-wrong",
      ),
      "packages/skill/skill-child/child-guide/SKILL.md": "# Child\n",
      "packages/skill/skill-child/.claude-plugin/plugin.json": manifest(
        "xonovex-skill-child",
        ["xonovex-skill-wrong"],
        "claude",
      ),
      "packages/skill/skill-child/.codex-plugin/plugin.json": manifest(
        "xonovex-skill-child",
        ["xonovex-skill-wrong"],
      ),
    });
    const report = makeSink();

    checkSkillDependencies(repo, report);

    expect(report.fails).toContain(
      "skill dependencies: manifests in packages/skill/skill-base must be named xonovex-skill-base",
    );
    expect(report.fails).toContain(
      "skill dependencies: xonovex-skill-child depends on xonovex-skill-wrong but does not name **base-guide** in its guidance",
    );
  });

  it("rejects stale and mismatched manifest skill paths", () => {
    const repo = makeRepo({
      "packages/skill/skill-renamed/renamed-guide/SKILL.md": "# Renamed\n",
      "packages/skill/skill-renamed/.claude-plugin/plugin.json": manifest(
        "xonovex-skill-renamed",
        [],
        "claude",
        "./old-guide",
      ),
      "packages/skill/skill-renamed/.codex-plugin/plugin.json": manifest(
        "xonovex-skill-renamed",
        [],
        "codex",
        "./renamed-guide",
      ),
    });
    const report = makeSink();

    checkSkillDependencies(repo, report);

    expect(report.fails).toContain(
      "skill packaging: manifest skill paths differ for xonovex-skill-renamed",
    );
    expect(report.fails).toContain(
      "skill packaging: Claude manifest in packages/skill/skill-renamed must point directly to ./renamed-guide",
    );
  });
});
