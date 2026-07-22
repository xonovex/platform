import {readFileSync} from "node:fs";
import {resolve} from "node:path";
import {describe, expect, it} from "vitest";
import {catalogEntries, selectEvalMatrix} from "./list-eval-matrix.mjs";

const entries = ["alpha", "beta", "gamma", "omega"].map((name) => ({
  package: name,
  project: `skill-${name}`,
}));

describe("selectEvalMatrix", () => {
  it("selects changed skills and expands shared evaluator changes", () => {
    expect(
      selectEvalMatrix({
        entries,
        changedFiles: ["packages/skill/skill-beta/beta-guide/SKILL.md"],
        limit: 12,
        offset: 0,
      }),
    ).toEqual([entries[1]]);
    expect(
      selectEvalMatrix({
        entries,
        changedFiles: [
          "packages/script/script-moon-skill-eval-triggers/src/evaluate.ts",
        ],
        limit: 2,
        offset: 0,
      }),
    ).toEqual(entries.slice(0, 2));
  });

  it("rotates bounded scheduled slices", () => {
    expect(
      selectEvalMatrix({
        entries,
        changedFiles: undefined,
        limit: 2,
        offset: 3,
      }),
    ).toEqual([entries[3], entries[0]]);
  });

  it("wires scheduled runs to a three-month catalog rotation", () => {
    const workflow = readFileSync(
      resolve(
        import.meta.dirname,
        "../../../../../.github/workflows/skill-evals.yml",
      ),
      "utf8",
    );

    expect(workflow).toMatch(
      /GITHUB_EVENT_NAME\}" == "schedule" \|\| \( "\$\{GITHUB_EVENT_NAME\}" == "workflow_dispatch" && "\$\{DISPATCH_SCOPE\}" == "rotated" \)/,
    );
    expect(workflow).toContain("limit=32");
    expect(workflow).toContain(
      'if [[ "${GITHUB_EVENT_NAME}" == "pull_request" ]]',
    );
    expect(workflow).toContain("limit=8");
    expect(workflow).toContain('offset="$(((10#$(date -u +%m) - 1) * 32))"');
    expect(workflow).toContain("-name invalid-run.json");
    expect(workflow).toContain("SKILL_ROUTING_OWNERS:");
    expect(workflow).toContain("SKILL_ROUTING_RUNS:");

    const catalog = catalogEntries(resolve(import.meta.dirname, "../../.."));
    const selected = new Set(
      [0, 32, 64].flatMap((offset) =>
        selectEvalMatrix({
          entries: catalog,
          changedFiles: undefined,
          limit: 32,
          offset,
        }).map(({project}) => project),
      ),
    );
    expect(selected.size).toBe(catalog.length);
  });
});
