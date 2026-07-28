import {mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {dirname, join} from "node:path";
import {type LinkReport} from "@xonovex/script-moon-skill-catalog-common/reference-file-links";
import {afterEach, describe, expect, it} from "vitest";
import {
  checkCrossPackageLinks,
  checkMarkdownFilesForCrossPackageLinks,
  checkNamedSkillHandoffs,
  checkSkillDependencies,
} from "./cross-package-links.js";

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
  const generatedPackageFiles = Object.fromEntries(
    Object.entries(files).flatMap(([path, content]) => {
      const match =
        /^(packages\/skill\/skill-[^/]+)\/\.claude-plugin\/plugin\.json$/u.exec(
          path,
        );
      if (match?.[1] === undefined) return [];
      const packagePath = `${match[1]}/package.json`;
      if (files[packagePath] !== undefined) return [];
      const plugin = JSON.parse(content) as {
        readonly dependencies?: readonly string[];
        readonly name: string;
      };
      const dependencies = Object.fromEntries(
        (plugin.dependencies ?? []).map((dependency) => [
          `@xonovex/${dependency.replace(/^xonovex-/u, "")}`,
          "7.0.0",
        ]),
      );
      return [
        [
          packagePath,
          JSON.stringify({
            name: `@xonovex/${plugin.name.replace(/^xonovex-/u, "")}`,
            version: "7.0.0",
            dependencies,
          }),
        ],
      ];
    }),
  );
  for (const [rel, content] of Object.entries({
    ...generatedPackageFiles,
    ...files,
  })) {
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

const commandNames = ["create", "review"] as const;

const commandFiles = (
  overrides: Readonly<Record<string, string>> = {},
): Record<string, string> =>
  Object.fromEntries(
    commandNames.map((name) => [
      `packages/command/command-test/commands/${name}.md`,
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
      "packages/command/command-test/commands/foo.md":
        "See [actors](../../../skill/skill-x/x-guide/references/actors.md).",
    });
    const source = join(repo, "packages/command/command-test/commands/foo.md");
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
      "packages/command/command-test/commands/foo.md":
        "See [actors](../../../skill/skill-x/x-guide/references/actors.md).",
    });
    const source = join(repo, "packages/command/command-test/commands/foo.md");
    const report = makeSink();
    const counts = checkMarkdownFilesForCrossPackageLinks(
      [source],
      repo,
      report,
    );
    expect(counts.broken).toBe(1);
    expect(report.fails).toHaveLength(1);
    expect(report.fails[0]).toContain(
      "packages/command/command-test/commands/foo.md",
    );
    expect(report.fails[0]).toContain(
      "../../../skill/skill-x/x-guide/references/actors.md",
    );
  });

  it("ignores an intra-package link, even a broken one", () => {
    const repo = makeRepo({
      "packages/command/command-test/docs/a.md": "See [b](./missing.md).",
    });
    const source = join(repo, "packages/command/command-test/docs/a.md");
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
      "packages/command/command-test/docs/skips.md": [
        "[a](https://example.com)",
        "[b](mailto:x@y.z)",
        "[c](../../../skill/<topic>.md)",
        "[d](../../../skill/{topic}.md)",
        "[e](…/pull/PR)",
        "[f](#anchor)",
      ].join("\n"),
    });
    const source = join(repo, "packages/command/command-test/docs/skips.md");
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
      "packages/command/command-test/docs/a.md":
        "See [g](../../../skill/skill-y/y-guide/references/g.md#section).",
    });
    const source = join(repo, "packages/command/command-test/docs/a.md");
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
  it("scans command packages and retained planning references", () => {
    const repo = makeRepo({
      ...commandFiles(),
      "packages/command/command-test/README.md":
        "See the [command model](../../diagram/diagram-command/command-model.png).\n",
      "packages/command/command-test/docs/invocation.md":
        "See the [submission boundary](../../../agent/agent-operator-go/README.md).\n",
      "packages/diagram/diagram-command/command-model.png": "image",
      "packages/agent/agent-operator-go/README.md": "# Operator\n",
      "packages/skill/skill-plan/plan-guide/SKILL.md":
        "# Plan\nSee [create](references/create.md).\n",
      "packages/skill/skill-plan/plan-guide/references/create.md": "# create\n",
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

  it("requires package-derived plugin names", () => {
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
  });

  it("requires plugin hard dependencies in the npm and Moon graph", () => {
    const repo = makeRepo({
      "packages/skill/skill-base/base-guide/SKILL.md": "# Base\n",
      "packages/skill/skill-base/.claude-plugin/plugin.json": manifest(
        "xonovex-skill-base",
        [],
        "claude",
      ),
      "packages/skill/skill-base/.codex-plugin/plugin.json":
        manifest("xonovex-skill-base"),
      "packages/skill/skill-child/package.json": JSON.stringify({
        name: "@xonovex/skill-child",
        version: "7.0.0",
      }),
      "packages/skill/skill-child/child-guide/SKILL.md":
        "# Child\nLoad **base-guide**.\n",
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

    expect(report.fails).toContain(
      "skill dependencies: xonovex-skill-child declares xonovex-skill-base in its plugin manifests but omits @xonovex/skill-base from package.json",
    );
  });

  it("requires exact bidirectional npm and plugin dependency parity", () => {
    const repo = makeRepo({
      "packages/skill/skill-base/base-guide/SKILL.md": "# Base\n",
      "packages/skill/skill-base/.claude-plugin/plugin.json": manifest(
        "xonovex-skill-base",
        [],
        "claude",
      ),
      "packages/skill/skill-base/.codex-plugin/plugin.json":
        manifest("xonovex-skill-base"),
      "packages/skill/skill-child/package.json": JSON.stringify({
        name: "@xonovex/skill-child",
        version: "7.0.0",
        dependencies: {
          "@xonovex/skill-base": "^7.0.0",
        },
      }),
      "packages/skill/skill-child/child-guide/SKILL.md": "# Child\n",
      "packages/skill/skill-child/.claude-plugin/plugin.json": manifest(
        "xonovex-skill-child",
        [],
        "claude",
      ),
      "packages/skill/skill-child/.codex-plugin/plugin.json": manifest(
        "xonovex-skill-child",
      ),
    });
    const report = makeSink();

    checkSkillDependencies(repo, report);

    expect(report.fails).toContain(
      "skill dependencies: xonovex-skill-child declares @xonovex/skill-base in package.json but omits xonovex-skill-base from its plugin manifests",
    );

    const pinnedRepo = makeRepo({
      "packages/skill/skill-base/base-guide/SKILL.md": "# Base\n",
      "packages/skill/skill-base/.claude-plugin/plugin.json": manifest(
        "xonovex-skill-base",
        [],
        "claude",
      ),
      "packages/skill/skill-base/.codex-plugin/plugin.json":
        manifest("xonovex-skill-base"),
      "packages/skill/skill-child/package.json": JSON.stringify({
        name: "@xonovex/skill-child",
        version: "7.0.0",
        dependencies: {
          "@xonovex/skill-base": "^7.0.0",
        },
      }),
      "packages/skill/skill-child/child-guide/SKILL.md":
        "# Child\nLoad **base-guide**.\n",
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
    const pinnedReport = makeSink();

    checkSkillDependencies(pinnedRepo, pinnedReport);

    expect(pinnedReport.fails).toContain(
      "skill dependencies: xonovex-skill-child pins @xonovex/skill-base@^7.0.0; expected exact installed version 7.0.0",
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

  it("rejects a dependent skill reference that substantially duplicates its dependency", () => {
    const duplicated = Array.from(
      {length: 40},
      (_, index) =>
        `shared concept phrase number ${String(index)} stays identical`,
    ).join(" ");
    const repo = makeRepo({
      "packages/skill/skill-base/base-guide/SKILL.md": "# Base\n",
      "packages/skill/skill-base/base-guide/references/concept.md": duplicated,
      "packages/skill/skill-base/.claude-plugin/plugin.json": manifest(
        "xonovex-skill-base",
        [],
        "claude",
      ),
      "packages/skill/skill-base/.codex-plugin/plugin.json":
        manifest("xonovex-skill-base"),
      "packages/skill/skill-child/child-guide/SKILL.md":
        "# Child\nUse **base-guide**.\n",
      "packages/skill/skill-child/child-guide/references/concept-copy.md":
        duplicated,
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

    expect(report.fails).toContainEqual(
      expect.stringContaining(
        "skill ownership: xonovex-skill-child child-guide/references/concept-copy.md duplicates xonovex-skill-base base-guide/references/concept.md",
      ),
    );
  });
});
