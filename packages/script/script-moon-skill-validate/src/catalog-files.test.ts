import {
  mkdirSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach, describe, expect, it} from "vitest";
import {checkCatalogFiles} from "./catalog-files.js";

const tempDirectories: string[] = [];

const makeSkill = (): string => {
  const skillDir = mkdtempSync(join(tmpdir(), "skill-catalog-files-"));
  tempDirectories.push(skillDir);
  writeFileSync(
    join(skillDir, "evals.json"),
    JSON.stringify({
      skill_name: "example-guide",
      tier: "moderate",
      evals: [1, 2, 3].map((id) => ({
        id,
        prompt: `prompt ${String(id)}`,
        assertions: ["observable assertion"],
      })),
    }),
  );
  writeFileSync(
    join(skillDir, "eval-queries.json"),
    JSON.stringify(
      [true, false].flatMap((shouldTrigger) =>
        Array.from({length: 8}, (_, index) => ({
          query: `${shouldTrigger ? "positive" : "negative"} ${String(index)}`,
          should_trigger: shouldTrigger,
          split: index < 5 ? "train" : "validation",
        })),
      ),
    ),
  );
  writeFileSync(
    join(skillDir, "SOURCES.md"),
    "# Sources\n\n## Primary\n\n- **URLs:**\n  - https://example.com/docs\n- **Last reviewed:** 2026-07-19\n",
  );
  return skillDir;
};

afterEach(() => {
  for (const directory of tempDirectories) {
    rmSync(directory, {recursive: true, force: true});
  }
  tempDirectories.length = 0;
});

describe("checkCatalogFiles", () => {
  it("accepts complete output, trigger, and source evidence", () => {
    const skillDir = makeSkill();

    const report = checkCatalogFiles(skillDir, {name: "example-guide"});

    expect(report.errors).toEqual([]);
    expect(report.passes).toHaveLength(3);
  });

  it("rejects missing files and underspecified routing fixtures", () => {
    const skillDir = makeSkill();
    rmSync(join(skillDir, "evals.json"));
    rmSync(join(skillDir, "SOURCES.md"));
    writeFileSync(
      join(skillDir, "eval-queries.json"),
      JSON.stringify([
        {query: "only one", should_trigger: true},
        {query: "only one", should_trigger: false},
      ]),
    );

    const report = checkCatalogFiles(skillDir, {name: "example-guide"});

    expect(report.errors).toContain("catalog: missing evals.json");
    expect(report.errors).toContain("catalog: missing SOURCES.md");
    expect(report.errors).toContain(
      "catalog: duplicate trigger query 'only one'",
    );
    expect(report.errors).toContain(
      "catalog: trigger evals need at least 8 positive and 8 negative queries (found 1/1)",
    );
  });

  it("requires capability metadata when a skill bundles scripts", () => {
    const skillDir = makeSkill();
    mkdirSync(join(skillDir, "scripts"));
    writeFileSync(join(skillDir, "scripts", "probe.sh"), "#!/bin/sh\n");

    const report = checkCatalogFiles(skillDir, {name: "example-guide"});

    expect(report.errors).toContain(
      "catalog: scripted skills need compatibility runtime/network metadata",
    );
    expect(report.errors).toContain(
      "catalog: scripted skills need a non-empty allowed-tools policy",
    );
  });

  it("rejects generated routing boilerplate", () => {
    const skillDir = makeSkill();
    const path = join(skillDir, "eval-queries.json");
    const queries = JSON.parse(readFileSync(path, "utf8")) as {
      query: string;
    }[];
    const query =
      "I'm reviewing `src/widgets` in a Example project. The happy path works, but widgets is unclear. Identify one realistic failure case.";
    const first = queries[0];
    if (first === undefined) throw new Error("fixture has no trigger queries");
    first.query = query;
    writeFileSync(path, JSON.stringify(queries));

    const report = checkCatalogFiles(skillDir, {name: "example-guide"});

    expect(report.errors).toContain(
      `catalog: generic trigger eval query must be replaced: '${query}'`,
    );
  });

  it.each([
    "A review comment on `work/widgets/` says the widgets change handles the normal fixture but leaves the failure path undefined.",
    "quick pre-merge sanity check: after the widgets change under `work/widgets/`, a clean checkout behaves differently from my local run.",
    "The clean Linux CI job fails only for the minimal widgets fixture under `work/widgets/`, while the full local fixture passes.",
  ])("rejects catalog-completion scenario templates", (query) => {
    const skillDir = makeSkill();
    const path = join(skillDir, "eval-queries.json");
    const queries = JSON.parse(readFileSync(path, "utf8")) as {
      query: string;
    }[];
    const first = queries[0];
    if (first === undefined) throw new Error("fixture has no trigger queries");
    first.query = query;
    writeFileSync(path, JSON.stringify(queries));

    const report = checkCatalogFiles(skillDir, {name: "example-guide"});

    expect(report.errors).toContain(
      `catalog: generic trigger eval query must be replaced: '${query}'`,
    );
  });

  it("rejects generated near misses concentrated in one sibling", () => {
    const skillDir = makeSkill();
    const path = join(skillDir, "eval-queries.json");
    const queries = JSON.parse(readFileSync(path, "utf8")) as {
      rationale?: string;
      should_trigger: boolean;
    }[];
    for (const query of queries.filter(({should_trigger}) => !should_trigger)) {
      query.rationale = "near miss owned by one-guide";
    }
    writeFileSync(path, JSON.stringify(queries));

    const report = checkCatalogFiles(skillDir, {name: "example-guide"});

    expect(report.errors).toContain(
      "catalog: generated negative routes need at least 3 sibling owners (found 1)",
    );
  });

  it("requires a source version for a version-pinned skill", () => {
    const skillDir = makeSkill();

    const report = checkCatalogFiles(skillDir, {
      name: "example-guide",
      description: "Use when editing Example 2.4+ projects.",
    });

    expect(report.errors).toContain(
      "catalog: version-pinned skill needs a Version field in SOURCES.md",
    );
  });
});
