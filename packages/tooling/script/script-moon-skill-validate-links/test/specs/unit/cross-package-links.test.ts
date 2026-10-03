import {join} from "node:path";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {memoryFileSystem} from "@xonovex/script-moon-common/file-system-memory";
import {type LinkReport} from "@xonovex/script-moon-skill-catalog-common/reference-file-links";
import {describe, expect, it} from "vitest";
import {
  checkCrossPackageLinks,
  checkMarkdownFilesForCrossPackageLinks,
  checkNamedSkillHandoffs,
  checkSkillDependencies,
} from "../../../src/cross-package-links.js";

const REPO = "/repo";

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

// Build a repo in memory whose files sit at the given repo-relative paths.
const makeRepo = (files: Record<string, string>): FileSystem => {
  const generatedPackageFiles = Object.fromEntries(
    Object.entries(files).flatMap(([path, content]) => {
      const match =
        /^(packages\/runtime\/plugin\/plugin-[^/]+)\/\.claude-plugin\/plugin\.json$/u.exec(
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
          `@xonovex/${dependency.replace(/^xonovex-/u, "plugin-")}`,
          "7.0.0",
        ]),
      );
      return [
        [
          packagePath,
          JSON.stringify({
            name: `@xonovex/${plugin.name.replace(/^xonovex-/u, "plugin-")}`,
            version: "7.0.0",
            dependencies,
          }),
        ],
      ];
    }),
  );
  return memoryFileSystem({
    files: Object.fromEntries(
      Object.entries({...generatedPackageFiles, ...files}).map(
        ([rel, content]) => [join(REPO, rel), content],
      ),
    ),
  });
};

const manifest = (
  name: string,
  dependencies: readonly string[] = [],
  _kind: "claude" | "codex" = "codex",
  skill = "./skills/",
): string =>
  JSON.stringify({
    name,
    dependencies,
    skills: skill,
  });

const commandNames = ["create", "review"] as const;

const commandFiles = (
  overrides: Readonly<Record<string, string>> = {},
): Record<string, string> =>
  Object.fromEntries(
    commandNames.map((name) => [
      `packages/runtime/plugin/plugin-test/commands/${name}.md`,
      overrides[name] ?? `# ${name}\n`,
    ]),
  );

describe("checkMarkdownFilesForCrossPackageLinks", () => {
  it("resolves a boundary-crossing link to an existing target", () => {
    const fs = makeRepo({
      "packages/runtime/plugin/plugin-x/skills/x-guide/references/actors.md":
        "# Actors\n",
      "packages/runtime/plugin/plugin-test/commands/foo.md":
        "See [actors](../../plugin-x/skills/x-guide/references/actors.md).",
    });
    const source = join(
      REPO,
      "packages/runtime/plugin/plugin-test/commands/foo.md",
    );
    const report = makeSink();
    const counts = checkMarkdownFilesForCrossPackageLinks(
      [source],
      REPO,
      report,
      fs,
    );
    expect(counts).toEqual({resolved: 1, broken: 0});
    expect(report.fails).toEqual([]);
  });

  it("fails a boundary-crossing link whose target is missing, naming the source and link", () => {
    const fs = makeRepo({
      "packages/runtime/plugin/plugin-test/commands/foo.md":
        "See [actors](../../plugin-x/skills/x-guide/references/actors.md).",
    });
    const source = join(
      REPO,
      "packages/runtime/plugin/plugin-test/commands/foo.md",
    );
    const report = makeSink();
    const counts = checkMarkdownFilesForCrossPackageLinks(
      [source],
      REPO,
      report,
      fs,
    );
    expect(counts.broken).toBe(1);
    expect(report.fails).toHaveLength(1);
    expect(report.fails[0]).toContain(
      "packages/runtime/plugin/plugin-test/commands/foo.md",
    );
    expect(report.fails[0]).toContain(
      "../../plugin-x/skills/x-guide/references/actors.md",
    );
  });

  it("ignores link-shaped text inside a fenced block or an inline span", () => {
    const fs = makeRepo({
      "packages/runtime/plugin/plugin-test/commands/foo.md": [
        "```c",
        "  m->listeners[i](../../plugin-x/skills/x-guide/references/gone.md);",
        "```",
        "",
        "A markdown hyperlink `[name](../../plugin-x/skills/x-guide/nope.md)`.",
      ].join("\n"),
    });
    const source = join(
      REPO,
      "packages/runtime/plugin/plugin-test/commands/foo.md",
    );
    const report = makeSink();

    const counts = checkMarkdownFilesForCrossPackageLinks(
      [source],
      REPO,
      report,
      fs,
    );

    expect(counts).toEqual({resolved: 0, broken: 0});
    expect(report.fails).toEqual([]);
  });

  it("ignores an intra-package link, even a broken one", () => {
    const fs = makeRepo({
      "packages/runtime/plugin/plugin-test/docs/a.md": "See [b](./missing.md).",
    });
    const source = join(REPO, "packages/runtime/plugin/plugin-test/docs/a.md");
    const report = makeSink();
    const counts = checkMarkdownFilesForCrossPackageLinks(
      [source],
      REPO,
      report,
      fs,
    );
    expect(counts).toEqual({resolved: 0, broken: 0});
    expect(report.fails).toEqual([]);
  });

  it("skips external, placeholder, ellipsis, and anchor forms", () => {
    const fs = makeRepo({
      "packages/runtime/plugin/plugin-test/docs/skips.md": [
        "[a](https://example.com)",
        "[b](mailto:x@y.z)",
        "[c](../../../skill/<topic>.md)",
        "[d](../../../skill/{topic}.md)",
        "[e](…/pull/PR)",
        "[f](#anchor)",
      ].join("\n"),
    });
    const source = join(
      REPO,
      "packages/runtime/plugin/plugin-test/docs/skips.md",
    );
    const report = makeSink();
    const counts = checkMarkdownFilesForCrossPackageLinks(
      [source],
      REPO,
      report,
      fs,
    );
    expect(counts).toEqual({resolved: 0, broken: 0});
    expect(report.fails).toEqual([]);
  });

  it("resolves a boundary-crossing link that carries an in-page fragment", () => {
    const fs = makeRepo({
      "packages/runtime/plugin/plugin-y/skills/y-guide/references/g.md":
        "# G\n",
      "packages/runtime/plugin/plugin-test/docs/a.md":
        "See [g](../../plugin-y/skills/y-guide/references/g.md#section).",
    });
    const source = join(REPO, "packages/runtime/plugin/plugin-test/docs/a.md");
    const report = makeSink();
    const counts = checkMarkdownFilesForCrossPackageLinks(
      [source],
      REPO,
      report,
      fs,
    );
    expect(counts).toEqual({resolved: 1, broken: 0});
    expect(report.fails).toEqual([]);
  });
});

describe("checkCrossPackageLinks", () => {
  it("scans command packages and retained planning references", () => {
    const fs = makeRepo({
      ...commandFiles(),
      "packages/runtime/plugin/plugin-test/README.md":
        "See the [command model](../../../asset/asset-diagrams/command-model.png).\n",
      "packages/runtime/plugin/plugin-test/docs/invocation.md":
        "See the [submission boundary](../../../../sandbox/operator/agent-operator-go/README.md).\n",
      "packages/asset/asset-diagrams/command-model.png": "image",
      "packages/sandbox/operator/agent-operator-go/README.md": "# Operator\n",
      "packages/runtime/plugin/plugin-workflow/skills/plan-guide/SKILL.md":
        "# Plan\nSee [create](references/create.md).\n",
      "packages/runtime/plugin/plugin-workflow/skills/plan-guide/references/create.md":
        "# create\n",
      "packages/runtime/plugin/plugin-workflow/.claude-plugin/plugin.json":
        manifest("xonovex-workflow", [], "claude"),
      "packages/runtime/plugin/plugin-workflow/.codex-plugin/plugin.json":
        manifest("xonovex-workflow"),
    });
    const report = makeSink();
    checkCrossPackageLinks(REPO, report, fs);
    expect(report.fails).toEqual([]);
    expect(report.passes).toContain("cross-package links: 2/2 link(s) resolve");
  });

  it("fails when a discovered SKILL.md points at a moved contract", () => {
    const fs = makeRepo({
      "packages/runtime/plugin/plugin-x/skills/x-guide/SKILL.md":
        "# X\nSee [g](../../../plugin-y/skills/y-guide/references/g.md).\n",
      "packages/runtime/plugin/plugin-x/.claude-plugin/plugin.json": manifest(
        "xonovex-x",
        [],
        "claude",
      ),
      "packages/runtime/plugin/plugin-x/.codex-plugin/plugin.json":
        manifest("xonovex-x"),
    });
    const report = makeSink();
    checkCrossPackageLinks(REPO, report, fs);
    expect(report.passes).not.toContain(
      "cross-package links: 1/1 link(s) resolve",
    );
    expect(report.fails).toHaveLength(1);
    expect(report.fails[0]).toContain(
      "packages/runtime/plugin/plugin-x/skills/x-guide/SKILL.md",
    );
  });
});

describe("checkNamedSkillHandoffs", () => {
  it("resolves catalog guide names and rejects missing handoffs", () => {
    const fs = makeRepo({
      "packages/runtime/plugin/plugin-x/skills/x-guide/SKILL.md":
        "Use **known-guide** and **missing-guide**.\n",
    });
    const source = join(
      REPO,
      "packages/runtime/plugin/plugin-x/skills/x-guide/SKILL.md",
    );
    const report = makeSink();

    const counts = checkNamedSkillHandoffs(
      [source],
      new Set(["known-guide"]),
      REPO,
      report,
      fs,
    );

    expect(counts).toEqual({resolved: 1, broken: 1});
    expect(report.fails).toContain(
      "skill handoffs: packages/runtime/plugin/plugin-x/skills/x-guide/SKILL.md names missing **missing-guide**",
    );
  });
});

describe("checkSkillDependencies", () => {
  it("accepts matching acyclic manifest pairs", () => {
    const fs = makeRepo({
      "packages/runtime/plugin/plugin-base/skills/base-guide/SKILL.md":
        "# Base\n",
      "packages/runtime/plugin/plugin-base/.claude-plugin/plugin.json":
        manifest("xonovex-base", [], "claude"),
      "packages/runtime/plugin/plugin-base/.codex-plugin/plugin.json":
        manifest("xonovex-base"),
      "packages/runtime/plugin/plugin-child/skills/child-guide/SKILL.md":
        "# Child\nUse **base-guide**.\n",
      "packages/runtime/plugin/plugin-child/.claude-plugin/plugin.json":
        manifest("xonovex-child", ["xonovex-base"], "claude"),
      "packages/runtime/plugin/plugin-child/.codex-plugin/plugin.json":
        manifest("xonovex-child", ["xonovex-base"]),
    });
    const report = makeSink();

    checkSkillDependencies(REPO, report, fs);

    expect(report.fails).toEqual([]);
    expect(report.passes).toContain(
      "skill dependencies: 2 manifest pair(s) agree with no dangling dependencies or cycles",
    );
  });

  it("rejects mismatched, dangling, and cyclic dependencies", () => {
    const fs = makeRepo({
      "packages/runtime/plugin/plugin-a/.claude-plugin/plugin.json": manifest(
        "xonovex-a",
        ["xonovex-b", "xonovex-missing", "xonovex-other"],
        "claude",
      ),
      "packages/runtime/plugin/plugin-a/.codex-plugin/plugin.json": manifest(
        "xonovex-a",
        ["xonovex-b", "xonovex-missing"],
      ),
      "packages/runtime/plugin/plugin-b/.claude-plugin/plugin.json": manifest(
        "xonovex-b",
        ["xonovex-a"],
        "claude",
      ),
      "packages/runtime/plugin/plugin-b/.codex-plugin/plugin.json": manifest(
        "xonovex-b",
        ["xonovex-a"],
      ),
    });
    const report = makeSink();

    checkSkillDependencies(REPO, report, fs);

    expect(report.fails).toContain(
      "skill dependencies: manifest dependencies differ for xonovex-a",
    );
    expect(report.fails).toContain(
      "skill dependencies: xonovex-a depends on missing xonovex-missing",
    );
    expect(report.fails).toContain(
      "skill dependencies: dependency cycle xonovex-a → xonovex-b → xonovex-a",
    );
  });

  it("rejects a discovered skill that has neither manifest", () => {
    const fs = makeRepo({
      "packages/runtime/plugin/plugin-unpackaged/skills/unpackaged-guide/SKILL.md":
        "# Unpackaged\n",
    });
    const report = makeSink();

    checkSkillDependencies(REPO, report, fs);

    expect(report.fails).toContain(
      "skill dependencies: packages/runtime/plugin/plugin-unpackaged needs both Claude and Codex manifests",
    );
  });

  it("requires package-derived plugin names", () => {
    const fs = makeRepo({
      "packages/runtime/plugin/plugin-base/skills/base-guide/SKILL.md":
        "# Base\n",
      "packages/runtime/plugin/plugin-base/.claude-plugin/plugin.json":
        manifest("xonovex-wrong", [], "claude"),
      "packages/runtime/plugin/plugin-base/.codex-plugin/plugin.json":
        manifest("xonovex-wrong"),
      "packages/runtime/plugin/plugin-child/skills/child-guide/SKILL.md":
        "# Child\n",
      "packages/runtime/plugin/plugin-child/.claude-plugin/plugin.json":
        manifest("xonovex-child", ["xonovex-wrong"], "claude"),
      "packages/runtime/plugin/plugin-child/.codex-plugin/plugin.json":
        manifest("xonovex-child", ["xonovex-wrong"]),
    });
    const report = makeSink();

    checkSkillDependencies(REPO, report, fs);

    expect(report.fails).toContain(
      "skill dependencies: manifests in packages/runtime/plugin/plugin-base must be named xonovex-base",
    );
  });

  it("requires plugin hard dependencies in the npm and Moon graph", () => {
    const fs = makeRepo({
      "packages/runtime/plugin/plugin-base/skills/base-guide/SKILL.md":
        "# Base\n",
      "packages/runtime/plugin/plugin-base/.claude-plugin/plugin.json":
        manifest("xonovex-base", [], "claude"),
      "packages/runtime/plugin/plugin-base/.codex-plugin/plugin.json":
        manifest("xonovex-base"),
      "packages/runtime/plugin/plugin-child/package.json": JSON.stringify({
        name: "@xonovex/plugin-child",
        version: "7.0.0",
      }),
      "packages/runtime/plugin/plugin-child/skills/child-guide/SKILL.md":
        "# Child\nLoad **base-guide**.\n",
      "packages/runtime/plugin/plugin-child/.claude-plugin/plugin.json":
        manifest("xonovex-child", ["xonovex-base"], "claude"),
      "packages/runtime/plugin/plugin-child/.codex-plugin/plugin.json":
        manifest("xonovex-child", ["xonovex-base"]),
    });
    const report = makeSink();

    checkSkillDependencies(REPO, report, fs);

    expect(report.fails).toContain(
      "skill dependencies: xonovex-child declares xonovex-base in its plugin manifests but omits @xonovex/plugin-base from package.json",
    );
  });

  it("requires exact bidirectional npm and plugin dependency parity", () => {
    const fs = makeRepo({
      "packages/runtime/plugin/plugin-base/skills/base-guide/SKILL.md":
        "# Base\n",
      "packages/runtime/plugin/plugin-base/.claude-plugin/plugin.json":
        manifest("xonovex-base", [], "claude"),
      "packages/runtime/plugin/plugin-base/.codex-plugin/plugin.json":
        manifest("xonovex-base"),
      "packages/runtime/plugin/plugin-child/package.json": JSON.stringify({
        name: "@xonovex/plugin-child",
        version: "7.0.0",
        dependencies: {
          "@xonovex/plugin-base": "^7.0.0",
        },
      }),
      "packages/runtime/plugin/plugin-child/skills/child-guide/SKILL.md":
        "# Child\n",
      "packages/runtime/plugin/plugin-child/.claude-plugin/plugin.json":
        manifest("xonovex-child", [], "claude"),
      "packages/runtime/plugin/plugin-child/.codex-plugin/plugin.json":
        manifest("xonovex-child"),
    });
    const report = makeSink();

    checkSkillDependencies(REPO, report, fs);

    expect(report.fails).toContain(
      "skill dependencies: xonovex-child declares @xonovex/plugin-base in package.json but omits xonovex-base from its plugin manifests",
    );

    const pinnedRepo = makeRepo({
      "packages/runtime/plugin/plugin-base/skills/base-guide/SKILL.md":
        "# Base\n",
      "packages/runtime/plugin/plugin-base/.claude-plugin/plugin.json":
        manifest("xonovex-base", [], "claude"),
      "packages/runtime/plugin/plugin-base/.codex-plugin/plugin.json":
        manifest("xonovex-base"),
      "packages/runtime/plugin/plugin-child/package.json": JSON.stringify({
        name: "@xonovex/plugin-child",
        version: "7.0.0",
        dependencies: {
          "@xonovex/plugin-base": "^7.0.0",
        },
      }),
      "packages/runtime/plugin/plugin-child/skills/child-guide/SKILL.md":
        "# Child\nLoad **base-guide**.\n",
      "packages/runtime/plugin/plugin-child/.claude-plugin/plugin.json":
        manifest("xonovex-child", ["xonovex-base"], "claude"),
      "packages/runtime/plugin/plugin-child/.codex-plugin/plugin.json":
        manifest("xonovex-child", ["xonovex-base"]),
    });
    const pinnedReport = makeSink();

    checkSkillDependencies(REPO, pinnedReport, pinnedRepo);

    expect(pinnedReport.fails).toContain(
      "skill dependencies: xonovex-child pins @xonovex/plugin-base@^7.0.0; expected exact installed version 7.0.0",
    );
  });

  it("rejects stale and mismatched manifest skill paths", () => {
    const fs = makeRepo({
      "packages/runtime/plugin/plugin-renamed/skills/renamed-guide/SKILL.md":
        "# Renamed\n",
      "packages/runtime/plugin/plugin-renamed/.claude-plugin/plugin.json":
        manifest("xonovex-renamed", [], "claude", "./old-guide"),
      "packages/runtime/plugin/plugin-renamed/.codex-plugin/plugin.json":
        manifest("xonovex-renamed", [], "codex", "./renamed-guide"),
    });
    const report = makeSink();

    checkSkillDependencies(REPO, report, fs);

    expect(report.fails).toContain(
      "skill packaging: manifest skill paths differ for xonovex-renamed",
    );
    expect(report.fails).toContain(
      "skill packaging: Claude manifest in packages/runtime/plugin/plugin-renamed must point directly to ./skills/",
    );
  });

  it("rejects a dependent skill reference that substantially duplicates its dependency", () => {
    const duplicated = Array.from(
      {length: 40},
      (_, index) =>
        `shared concept phrase number ${String(index)} stays identical`,
    ).join(" ");
    const fs = makeRepo({
      "packages/runtime/plugin/plugin-base/skills/base-guide/SKILL.md":
        "# Base\n",
      "packages/runtime/plugin/plugin-base/skills/base-guide/references/concept.md":
        duplicated,
      "packages/runtime/plugin/plugin-base/.claude-plugin/plugin.json":
        manifest("xonovex-base", [], "claude"),
      "packages/runtime/plugin/plugin-base/.codex-plugin/plugin.json":
        manifest("xonovex-base"),
      "packages/runtime/plugin/plugin-child/skills/child-guide/SKILL.md":
        "# Child\nUse **base-guide**.\n",
      "packages/runtime/plugin/plugin-child/skills/child-guide/references/concept-copy.md":
        duplicated,
      "packages/runtime/plugin/plugin-child/.claude-plugin/plugin.json":
        manifest("xonovex-child", ["xonovex-base"], "claude"),
      "packages/runtime/plugin/plugin-child/.codex-plugin/plugin.json":
        manifest("xonovex-child", ["xonovex-base"]),
    });
    const report = makeSink();

    checkSkillDependencies(REPO, report, fs);

    expect(report.fails).toContainEqual(
      expect.stringContaining(
        "skill ownership: xonovex-child skills/child-guide/references/concept-copy.md duplicates xonovex-base skills/base-guide/references/concept.md",
      ),
    );
  });
});
