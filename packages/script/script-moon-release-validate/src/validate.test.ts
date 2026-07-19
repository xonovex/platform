import {execFileSync} from "node:child_process";
import {dirname, resolve} from "node:path";
import {fileURLToPath} from "node:url";
import {describe, expect, it} from "vitest";
import {
  blanketIdBlockFailures,
  createChecker,
  forbiddenClaims,
  markdownLinkTargets,
  offLegendClassifications,
  releaseWorkflowFailures,
  tableIds,
  tableShapeFailures,
} from "./validate.js";

describe("createChecker", () => {
  it("counts every check and collects only failures", () => {
    const checker = createChecker();
    checker.check(true, "passes");
    checker.check(false, "fails");
    checker.check(false, "also fails");
    expect(checker.checks()).toBe(3);
    expect(checker.failures()).toEqual(["fails", "also fails"]);
  });
});

describe("release validator", () => {
  it("validates the repository release inputs end to end", () => {
    const sourceDirectory = dirname(fileURLToPath(import.meta.url));
    const entrypoint = resolve(sourceDirectory, "../dist/src/index.js");

    const output = execFileSync(process.execPath, [entrypoint], {
      encoding: "utf8",
    });

    expect(output).toContain("Release validation passed:");
  });
});

describe("tableIds", () => {
  it("extracts prefixed ids from leading table cells only", () => {
    const content = [
      "| ID | Title |",
      "| --- | --- |",
      "| D-001 | first |",
      "| D-002 | mentions D-099 inline |",
      "| note | D-100 not in the leading cell |",
    ].join("\n");
    expect([...tableIds(content, "D-")]).toEqual(["D-001", "D-002"]);
  });
});

describe("markdownLinkTargets", () => {
  it("returns relative targets and skips external and anchor-only links", () => {
    const content =
      "[a](./local.md) [b](https://example.com) [c](mailto:x@y.z) " +
      "[d](#anchor) [e](../up.md#section)";
    expect(markdownLinkTargets(content)).toEqual(["./local.md", "../up.md"]);
  });
});

describe("forbiddenClaims", () => {
  it("matches overclaiming phrases and passes qualified statements", () => {
    const overclaim = "Installing a skill enforces the policy everywhere.";
    const qualified = "A skill documents the policy; enforcement is separate.";
    expect(forbiddenClaims.some((claim) => claim.test(overclaim))).toBe(true);
    expect(forbiddenClaims.some((claim) => claim.test(qualified))).toBe(false);
  });
});

describe("tableShapeFailures", () => {
  it("flags rows whose cell count differs from their header", () => {
    const good = "| a | b |\n| --- | --- |\n| 1 | 2 |";
    const bad = "| a | b |\n| --- | --- |\n| 1 | 2 | 3 |";
    expect(tableShapeFailures(good, "t")).toEqual([]);
    expect(tableShapeFailures(bad, "t")).toHaveLength(1);
  });
});

describe("offLegendClassifications", () => {
  it("flags classification cells outside the legend", () => {
    const legend = new Set(["platform-derived"]);
    const rows =
      "| D-001 | statement | platform-derived | S-A | why |\n" +
      "| D-002 | statement | governance synthesis | S-A | why |";
    expect(offLegendClassifications(rows, legend)).toEqual([
      "D-002: classification 'governance synthesis' is not in the legend",
    ]);
  });
});

describe("blanketIdBlockFailures", () => {
  it("flags a subplan whose every task row shares one ID block", () => {
    const blanket =
      "| subplan-01-x.md | 1 | intent | D-001 | C-001 | S-A |\n" +
      "| subplan-01-x.md | 2 | other | D-001 | C-001 | S-A |";
    const varied =
      "| subplan-02-y.md | 1 | intent | D-001 | C-001 | S-A |\n" +
      "| subplan-02-y.md | 2 | other | D-002 | C-001 | S-A |";
    expect(blanketIdBlockFailures(blanket)).toHaveLength(1);
    expect(blanketIdBlockFailures(varied)).toEqual([]);
    expect(blanketIdBlockFailures(blanket + "\n" + varied)).toHaveLength(1);
  });
});

describe("releaseWorkflowFailures", () => {
  it("accepts PR-only publishing with dispatch limited to dry runs", () => {
    const workflow = `
on:
  workflow_dispatch:
  pull_request:
jobs:
  release:
    steps:
      - name: Publish
        if: github.event_name == 'pull_request'
        run: publish
      - name: Publish (dry run)
        if: github.event_name == 'workflow_dispatch'
        run: dry-run
      - uses: example/report@sha
`;

    expect(releaseWorkflowFailures(workflow)).toEqual([]);
  });

  it("rejects tag and dispatch-controlled publishing", () => {
    const workflow = `
on:
  push:
    tags: ["v*"]
jobs:
  release:
    steps:
      - name: Publish
        if: inputs.dry_run != true
        run: publish
      - name: Publish (dry run)
        if: inputs.dry_run == true
        run: dry-run
      - uses: example/report@sha
`;

    expect(releaseWorkflowFailures(workflow)).toEqual([
      "release workflow must not publish from pushed tags",
      "release workflow must not derive publish authority from a dispatch input",
      "publish step must be restricted to the pull_request event",
      "manual dispatch must be restricted to the dry-run publish step",
    ]);
  });
});
