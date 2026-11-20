import {memoryOutput} from "@xonovex/script-moon-common/adapters/output-memory";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import {describe, expect, it} from "vitest";
import {main} from "../../../src/validate-links.js";
import {makeRepo, plugin, REPO} from "../../util/repository.js";

const run = (argv: readonly string[], fs: FileSystem) => {
  const output = memoryOutput();
  const code = main(argv, REPO, {fs, output});
  return {code, lines: output.stdoutLines()};
};

describe("moon-skill-validate-links", () => {
  it("renders help without scanning the repository", () => {
    const result = run(["--help"], makeRepo({}));

    expect(result.code).toBe(0);
    expect(result.lines[0]).toContain("usage: moon-skill-validate-links");
  });

  it("passes a repository whose links, handoffs and packages are valid", () => {
    const fs = makeRepo({
      ...plugin("base"),
      ...plugin(
        "child",
        ["xonovex-base"],
        "# Child\nUse **using-base**, see [base](../../../plugin-base/skills/using-base/SKILL.md).\n",
      ),
      "packages/runtime/plugin/plugin-child/README.md":
        "See the [operator](../../../sandbox/operator/agent-operator-go/README.md).\n",
      "packages/sandbox/operator/agent-operator-go/README.md": "# Operator\n",
    });

    expect(run([], fs)).toEqual({
      code: 0,
      lines: [
        "[PASS] cross-package links: 2/2 link(s) resolve",
        "[PASS] skill handoffs: 1/1 named handoff(s) resolve",
        "[PASS] skill dependencies: 2 manifest pair(s) agree with no dangling dependencies or cycles",
        "Result: PASS (links, packaging, and hard dependencies are valid)",
      ],
    });
  });

  it("fails a repository with a broken link and a missing handoff", () => {
    const fs = makeRepo({
      ...plugin(
        "x",
        [],
        "# X\nUse **using-gone**, see [g](../../../plugin-y/skills/using-y/references/g.md).\n",
      ),
    });

    expect(run([REPO], fs)).toEqual({
      code: 1,
      lines: [
        "[PASS] skill dependencies: 1 manifest pair(s) agree with no dangling dependencies or cycles",
        "[FAIL] cross-package links: broken link in packages/runtime/plugin/plugin-x/skills/using-x/SKILL.md → ../../../plugin-y/skills/using-y/references/g.md",
        "[FAIL] skill handoffs: packages/runtime/plugin/plugin-x/skills/using-x/SKILL.md names missing **using-gone**",
        "Result: FAIL (2 dependency or link error(s))",
      ],
    });
  });

  it("resolves a relative root against the working directory", () => {
    const fs = makeRepo({
      ...plugin(
        "x",
        [],
        "# X\nSee [r](../../../../../../README.md) and [l](./local.md).\n",
      ),
      "README.md": "# Root\n",
      "packages/runtime/plugin/plugin-x/skills/using-x/local.md": "# Local\n",
    });

    expect(run(["."], fs).lines[0]).toBe(
      "[PASS] cross-package links: 1/1 link(s) resolve",
    );
  });
});
