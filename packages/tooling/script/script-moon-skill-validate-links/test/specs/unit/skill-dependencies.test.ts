import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import {describe, expect, it} from "vitest";
import {scanCatalog} from "../../../src/catalog-scan.js";
import {checkSkillDependencies} from "../../../src/skill-dependencies.js";
import {makeRepo, manifest, plugin, REPO} from "../../util/repository.js";

// The failure messages of the dependency check over a scanned repository.
const failures = (fs: FileSystem): readonly string[] =>
  checkSkillDependencies(scanCatalog(REPO, fs).plugins, REPO)
    .filter(({level}) => level === "fail")
    .map(({message}) => message);

describe("checkSkillDependencies", () => {
  it("accepts matching acyclic manifest pairs", () => {
    const fs = makeRepo({
      "packages/runtime/plugin/plugin-base/skills/using-base/SKILL.md":
        "# Base\n",
      "packages/runtime/plugin/plugin-base/.claude-plugin/plugin.json":
        manifest("xonovex-base"),
      "packages/runtime/plugin/plugin-base/.codex-plugin/plugin.json":
        manifest("xonovex-base"),
      "packages/runtime/plugin/plugin-child/skills/using-child/SKILL.md":
        "# Child\nUse **using-base**.\n",
      "packages/runtime/plugin/plugin-child/.claude-plugin/plugin.json":
        manifest("xonovex-child", ["xonovex-base"]),
      "packages/runtime/plugin/plugin-child/.codex-plugin/plugin.json":
        manifest("xonovex-child", ["xonovex-base"]),
    });

    expect(checkSkillDependencies(scanCatalog(REPO, fs).plugins, REPO)).toEqual(
      [
        {
          level: "pass",
          message:
            "skill dependencies: 2 manifest pair(s) agree with no dangling dependencies or cycles",
        },
      ],
    );
  });

  it("rejects mismatched, dangling, and cyclic dependencies", () => {
    const fs = makeRepo({
      "packages/runtime/plugin/plugin-a/.claude-plugin/plugin.json": manifest(
        "xonovex-a",
        ["xonovex-b", "xonovex-missing", "xonovex-other"],
      ),
      "packages/runtime/plugin/plugin-a/.codex-plugin/plugin.json": manifest(
        "xonovex-a",
        ["xonovex-b", "xonovex-missing"],
      ),
      "packages/runtime/plugin/plugin-b/.claude-plugin/plugin.json": manifest(
        "xonovex-b",
        ["xonovex-a"],
      ),
      "packages/runtime/plugin/plugin-b/.codex-plugin/plugin.json": manifest(
        "xonovex-b",
        ["xonovex-a"],
      ),
    });

    expect(failures(fs)).toContain(
      "skill dependencies: manifest dependencies differ for xonovex-a",
    );
    expect(failures(fs)).toContain(
      "skill dependencies: xonovex-a depends on missing xonovex-missing",
    );
    expect(failures(fs)).toContain(
      "skill dependencies: dependency cycle xonovex-a → xonovex-b → xonovex-a",
    );
  });

  it("rejects a discovered skill that has neither manifest", () => {
    const fs = makeRepo({
      "packages/runtime/plugin/plugin-unpackaged/skills/using-unpackaged/SKILL.md":
        "# Unpackaged\n",
    });

    expect(failures(fs)).toContain(
      "skill dependencies: packages/runtime/plugin/plugin-unpackaged needs both Claude and Codex manifests",
    );
  });

  it("requires package-derived plugin names", () => {
    const fs = makeRepo({
      "packages/runtime/plugin/plugin-base/skills/using-base/SKILL.md":
        "# Base\n",
      "packages/runtime/plugin/plugin-base/.claude-plugin/plugin.json":
        manifest("xonovex-wrong"),
      "packages/runtime/plugin/plugin-base/.codex-plugin/plugin.json":
        manifest("xonovex-wrong"),
      "packages/runtime/plugin/plugin-child/skills/using-child/SKILL.md":
        "# Child\n",
      "packages/runtime/plugin/plugin-child/.claude-plugin/plugin.json":
        manifest("xonovex-child", ["xonovex-wrong"]),
      "packages/runtime/plugin/plugin-child/.codex-plugin/plugin.json":
        manifest("xonovex-child", ["xonovex-wrong"]),
    });

    expect(failures(fs)).toContain(
      "skill dependencies: manifests in packages/runtime/plugin/plugin-base must be named xonovex-base",
    );
  });

  it("requires plugin hard dependencies in the npm and Moon graph", () => {
    const fs = makeRepo({
      "packages/runtime/plugin/plugin-base/skills/using-base/SKILL.md":
        "# Base\n",
      "packages/runtime/plugin/plugin-base/.claude-plugin/plugin.json":
        manifest("xonovex-base"),
      "packages/runtime/plugin/plugin-base/.codex-plugin/plugin.json":
        manifest("xonovex-base"),
      "packages/runtime/plugin/plugin-child/package.json": JSON.stringify({
        name: "@xonovex/plugin-child",
        version: "7.0.0",
      }),
      "packages/runtime/plugin/plugin-child/skills/using-child/SKILL.md":
        "# Child\nLoad **using-base**.\n",
      "packages/runtime/plugin/plugin-child/.claude-plugin/plugin.json":
        manifest("xonovex-child", ["xonovex-base"]),
      "packages/runtime/plugin/plugin-child/.codex-plugin/plugin.json":
        manifest("xonovex-child", ["xonovex-base"]),
    });

    expect(failures(fs)).toContain(
      "skill dependencies: xonovex-child declares xonovex-base in its plugin manifests but omits @xonovex/plugin-base from package.json",
    );
  });

  it("requires exact bidirectional npm and plugin dependency parity", () => {
    const fs = makeRepo({
      "packages/runtime/plugin/plugin-base/skills/using-base/SKILL.md":
        "# Base\n",
      "packages/runtime/plugin/plugin-base/.claude-plugin/plugin.json":
        manifest("xonovex-base"),
      "packages/runtime/plugin/plugin-base/.codex-plugin/plugin.json":
        manifest("xonovex-base"),
      "packages/runtime/plugin/plugin-child/package.json": JSON.stringify({
        name: "@xonovex/plugin-child",
        version: "7.0.0",
        dependencies: {
          "@xonovex/plugin-base": "^7.0.0",
        },
      }),
      "packages/runtime/plugin/plugin-child/skills/using-child/SKILL.md":
        "# Child\n",
      "packages/runtime/plugin/plugin-child/.claude-plugin/plugin.json":
        manifest("xonovex-child"),
      "packages/runtime/plugin/plugin-child/.codex-plugin/plugin.json":
        manifest("xonovex-child"),
    });

    expect(failures(fs)).toContain(
      "skill dependencies: xonovex-child declares @xonovex/plugin-base in package.json but omits xonovex-base from its plugin manifests",
    );

    const pinnedRepo = makeRepo({
      "packages/runtime/plugin/plugin-base/skills/using-base/SKILL.md":
        "# Base\n",
      "packages/runtime/plugin/plugin-base/.claude-plugin/plugin.json":
        manifest("xonovex-base"),
      "packages/runtime/plugin/plugin-base/.codex-plugin/plugin.json":
        manifest("xonovex-base"),
      "packages/runtime/plugin/plugin-child/package.json": JSON.stringify({
        name: "@xonovex/plugin-child",
        version: "7.0.0",
        dependencies: {
          "@xonovex/plugin-base": "^7.0.0",
        },
      }),
      "packages/runtime/plugin/plugin-child/skills/using-child/SKILL.md":
        "# Child\nLoad **using-base**.\n",
      "packages/runtime/plugin/plugin-child/.claude-plugin/plugin.json":
        manifest("xonovex-child", ["xonovex-base"]),
      "packages/runtime/plugin/plugin-child/.codex-plugin/plugin.json":
        manifest("xonovex-child", ["xonovex-base"]),
    });

    expect(failures(pinnedRepo)).toContain(
      "skill dependencies: xonovex-child pins @xonovex/plugin-base@^7.0.0; expected exact installed version 7.0.0",
    );
  });

  it("rejects stale and mismatched manifest skill paths", () => {
    const fs = makeRepo({
      "packages/runtime/plugin/plugin-renamed/skills/using-renamed/SKILL.md":
        "# Renamed\n",
      "packages/runtime/plugin/plugin-renamed/.claude-plugin/plugin.json":
        manifest("xonovex-renamed", [], "./using-old"),
      "packages/runtime/plugin/plugin-renamed/.codex-plugin/plugin.json":
        manifest("xonovex-renamed", [], "./using-renamed"),
    });

    expect(failures(fs)).toContain(
      "skill packaging: manifest skill paths differ for xonovex-renamed",
    );
    expect(failures(fs)).toContain(
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
      "packages/runtime/plugin/plugin-base/skills/using-base/SKILL.md":
        "# Base\n",
      "packages/runtime/plugin/plugin-base/skills/using-base/references/concept.md":
        duplicated,
      "packages/runtime/plugin/plugin-base/.claude-plugin/plugin.json":
        manifest("xonovex-base"),
      "packages/runtime/plugin/plugin-base/.codex-plugin/plugin.json":
        manifest("xonovex-base"),
      "packages/runtime/plugin/plugin-child/skills/using-child/SKILL.md":
        "# Child\nUse **using-base**.\n",
      "packages/runtime/plugin/plugin-child/skills/using-child/references/concept-copy.md":
        duplicated,
      "packages/runtime/plugin/plugin-child/.claude-plugin/plugin.json":
        manifest("xonovex-child", ["xonovex-base"]),
      "packages/runtime/plugin/plugin-child/.codex-plugin/plugin.json":
        manifest("xonovex-child", ["xonovex-base"]),
    });

    expect(failures(fs)).toContainEqual(
      expect.stringContaining(
        "skill ownership: xonovex-child skills/using-child/references/concept-copy.md duplicates xonovex-base skills/using-base/references/concept.md",
      ),
    );
  });

  it("rejects manifests that do not parse or do not agree on a name", () => {
    const fs = makeRepo({
      ...plugin("base"),
      "packages/runtime/plugin/plugin-broken/package.json": "{}",
      "packages/runtime/plugin/plugin-broken/.claude-plugin/plugin.json": "{",
      "packages/runtime/plugin/plugin-broken/.codex-plugin/plugin.json":
        JSON.stringify({name: 7}),
      "packages/runtime/plugin/plugin-split/.claude-plugin/plugin.json":
        manifest("xonovex-split"),
      "packages/runtime/plugin/plugin-split/.codex-plugin/plugin.json":
        manifest("xonovex-other"),
      "packages/runtime/plugin/plugin-loose/package.json": "{}",
      "packages/runtime/plugin/plugin-loose/.claude-plugin/plugin.json":
        JSON.stringify({name: "xonovex-loose", dependencies: [1]}),
      "packages/runtime/plugin/plugin-loose/.codex-plugin/plugin.json":
        JSON.stringify({name: "xonovex-loose"}),
      "packages/runtime/plugin/plugin-twin/skills/using-twin/SKILL.md": "# T\n",
      "packages/runtime/plugin/plugin-twin/.claude-plugin/plugin.json":
        manifest("xonovex-base"),
      "packages/runtime/plugin/plugin-twin/.codex-plugin/plugin.json":
        manifest("xonovex-base"),
      "packages/runtime/plugin/readme-only/README.md": "# Not a plugin\n",
      "packages/runtime/plugin/plugin-assets/README.md": "# No skills\n",
    });

    expect(failures(fs)).toEqual([
      expect.stringMatching(
        /^skill dependencies: invalid JSON in packages\/runtime\/plugin\/plugin-broken\/\.claude-plugin\/plugin\.json: /u,
      ) as unknown,
      "skill dependencies: packages/runtime/plugin/plugin-broken/.codex-plugin/plugin.json needs a string name",
      "skill dependencies: packages/runtime/plugin/plugin-loose/.claude-plugin/plugin.json dependencies must be strings",
      "skill packaging: packages/runtime/plugin/plugin-loose/.codex-plugin/plugin.json skills must be a string",
      "skill dependencies: manifests in packages/runtime/plugin/plugin-split must be named xonovex-split",
      "skill dependencies: manifest names differ in packages/runtime/plugin/plugin-split (xonovex-split != xonovex-other)",
      "skill dependencies: manifests in packages/runtime/plugin/plugin-twin must be named xonovex-twin",
      "skill dependencies: duplicate plugin name xonovex-base",
    ]);
  });

  it("rejects a package without skills and unreadable package files", () => {
    const fs = makeRepo({
      ...plugin("base"),
      ...plugin("child", ["xonovex-base", "xonovex-ghost"]),
      ...plugin("ghost"),
      "packages/runtime/plugin/plugin-base/package.json": JSON.stringify({
        dependencies: {"@xonovex/plugin-x": 1},
      }),
      "packages/runtime/plugin/plugin-child/package.json": JSON.stringify({
        dependencies: {
          "@xonovex/plugin-base": "1.0.0",
          "@xonovex/plugin-ghost": "1.0.0",
        },
      }),
      "packages/runtime/plugin/plugin-ghost/package.json": "{",
      "packages/runtime/plugin/plugin-empty/.claude-plugin/plugin.json":
        manifest("xonovex-empty"),
      "packages/runtime/plugin/plugin-empty/.codex-plugin/plugin.json":
        manifest("xonovex-empty"),
    });

    expect(failures(fs)).toEqual([
      "skill packaging: packages/runtime/plugin/plugin-empty contains no skills",
      "skill dependencies: packages/runtime/plugin/plugin-base/package.json dependencies must map package names to versions",
      "skill dependencies: packages/runtime/plugin/plugin-base/package.json needs a string version",
      expect.stringMatching(
        /^skill dependencies: cannot read packages\/runtime\/plugin\/plugin-ghost\/package\.json: /u,
      ) as unknown,
      expect.stringMatching(
        /^skill dependencies: cannot read packages\/runtime\/plugin\/plugin-ghost\/package\.json: /u,
      ) as unknown,
    ]);
  });

  it("reads a package a plugin name points at but no directory holds as missing", () => {
    const fs = makeRepo({
      ...plugin("child", ["xonovex-base"]),
      "packages/runtime/plugin/plugin-other/skills/using-other/SKILL.md":
        "# Other\n",
      "packages/runtime/plugin/plugin-other/.claude-plugin/plugin.json":
        manifest("xonovex-base"),
      "packages/runtime/plugin/plugin-other/.codex-plugin/plugin.json":
        manifest("xonovex-base"),
    });

    expect(failures(fs)).toContain(
      "skill dependencies: cannot read packages/runtime/plugin/plugin-base/package.json: ENOENT: no such file or directory, open '/repo/packages/runtime/plugin/plugin-base/package.json'",
    );
  });
});
