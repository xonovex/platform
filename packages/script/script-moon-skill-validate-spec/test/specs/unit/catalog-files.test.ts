import {join} from "node:path";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {memoryFileSystem} from "@xonovex/script-moon-common/file-system-memory";
import {describe, expect, it} from "vitest";
import {checkCatalogFiles} from "../../../src/catalog-files.js";

const SKILL_DIR = "/skill";

// The evidence a complete skill ships. A case that needs one piece absent builds
// the tree without it rather than deleting it.
interface SkillParts {
  readonly evals?: boolean;
  readonly sources?: boolean;
}

const makeSkill = ({
  evals = true,
  sources = true,
}: SkillParts = {}): FileSystem => {
  const fs = memoryFileSystem({directories: [SKILL_DIR]});
  if (evals) {
    fs.writeFile(
      join(SKILL_DIR, "evals.json"),
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
  }
  fs.writeFile(
    join(SKILL_DIR, "eval-queries.json"),
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
  if (sources) {
    fs.writeFile(
      join(SKILL_DIR, "SOURCES.md"),
      "# Sources\n\n## Primary\n\n- **URLs:**\n  - https://example.com/docs\n- **Last reviewed:** 2026-07-19\n",
    );
  }
  return fs;
};

describe("checkCatalogFiles", () => {
  it("accepts complete output, trigger, and source evidence", () => {
    const fs = makeSkill();

    const report = checkCatalogFiles(SKILL_DIR, {name: "example-guide"}, fs);

    expect(report.errors).toEqual([]);
    expect(report.passes).toHaveLength(4);
  });

  it("rejects missing files and underspecified routing fixtures", () => {
    const fs = makeSkill({evals: false, sources: false});
    fs.writeFile(
      join(SKILL_DIR, "eval-queries.json"),
      JSON.stringify([
        {query: "only one", should_trigger: true},
        {query: "only one", should_trigger: false},
      ]),
    );

    const report = checkCatalogFiles(SKILL_DIR, {name: "example-guide"}, fs);

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
    const fs = makeSkill();
    fs.makeDirectory(join(SKILL_DIR, "scripts"));
    fs.writeFile(join(SKILL_DIR, "scripts", "probe.sh"), "#!/bin/sh\n");

    const report = checkCatalogFiles(SKILL_DIR, {name: "example-guide"}, fs);

    expect(report.errors).toContain(
      "catalog: scripted skills need compatibility runtime/network metadata",
    );
    expect(report.errors).toContain(
      "catalog: scripted skills need a non-empty allowed-tools policy",
    );
  });

  it("rejects TypeScript and MJS implementation files", () => {
    const fs = makeSkill();
    fs.makeDirectory(join(SKILL_DIR, "scripts"));
    fs.writeFile(join(SKILL_DIR, "scripts", "validate.ts"), "");
    fs.writeFile(join(SKILL_DIR, "scripts", "evaluate.mjs"), "");

    expect(
      checkCatalogFiles(SKILL_DIR, {name: "example-guide"}, fs).errors,
    ).toContain(
      "catalog: skills must not contain TypeScript or MJS implementation files: scripts/evaluate.mjs, scripts/validate.ts",
    );
  });

  it("rejects generated routing boilerplate", () => {
    const fs = makeSkill();
    const path = join(SKILL_DIR, "eval-queries.json");
    const queries = JSON.parse(fs.readText(path)) as {
      query: string;
    }[];
    const query =
      "I'm reviewing `src/widgets` in a Example project. The happy path works, but widgets is unclear. Identify one realistic failure case.";
    const first = queries[0];
    if (first === undefined) throw new Error("fixture has no trigger queries");
    first.query = query;
    fs.writeFile(path, JSON.stringify(queries));

    const report = checkCatalogFiles(SKILL_DIR, {name: "example-guide"}, fs);

    expect(report.errors).toContain(
      `catalog: generic trigger eval query must be replaced: '${query}'`,
    );
  });

  it.each([
    "A review comment on `work/widgets/` says the widgets change handles the normal fixture but leaves the failure path undefined.",
    "quick pre-merge sanity check: after the widgets change under `work/widgets/`, a clean checkout behaves differently from my local run.",
    "The clean Linux CI job fails only for the minimal widgets fixture under `work/widgets/`, while the full local fixture passes.",
  ])("rejects catalog-completion scenario templates", (query) => {
    const fs = makeSkill();
    const path = join(SKILL_DIR, "eval-queries.json");
    const queries = JSON.parse(fs.readText(path)) as {
      query: string;
    }[];
    const first = queries[0];
    if (first === undefined) throw new Error("fixture has no trigger queries");
    first.query = query;
    fs.writeFile(path, JSON.stringify(queries));

    const report = checkCatalogFiles(SKILL_DIR, {name: "example-guide"}, fs);

    expect(report.errors).toContain(
      `catalog: generic trigger eval query must be replaced: '${query}'`,
    );
  });

  it("rejects generated near misses concentrated in one sibling", () => {
    const fs = makeSkill();
    const path = join(SKILL_DIR, "eval-queries.json");
    const queries = JSON.parse(fs.readText(path)) as {
      rationale?: string;
      should_trigger: boolean;
    }[];
    for (const query of queries.filter(({should_trigger}) => !should_trigger)) {
      query.rationale = "near miss owned by one-guide";
    }
    fs.writeFile(path, JSON.stringify(queries));

    const report = checkCatalogFiles(SKILL_DIR, {name: "example-guide"}, fs);

    expect(report.errors).toContain(
      "catalog: generated negative routes need at least 3 sibling owners (found 1)",
    );
  });

  it("requires a source version for a version-pinned skill", () => {
    const fs = makeSkill();

    const report = checkCatalogFiles(
      SKILL_DIR,
      {
        name: "example-guide",
        description: "Use when editing Example 2.4+ projects.",
      },
      fs,
    );

    expect(report.errors).toContain(
      "catalog: version-pinned skill needs a Version field in SOURCES.md",
    );
  });

  it("requires content or repository drift evidence for versioned web sources", () => {
    const fs = makeSkill();
    fs.writeFile(
      join(SKILL_DIR, "SOURCES.md"),
      `# Sources

## Versioned docs
- **URL:** https://example.com/docs
- **Version:** 2.4.0
- **References:** all
- **Last reviewed:** 2026-07-22
`,
    );

    const report = checkCatalogFiles(SKILL_DIR, {name: "example-guide"}, fs);

    expect(report.errors).toContain(
      "catalog: versioned web source 'Versioned docs' needs Content SHA256 or Checkout + Commit + Watch drift fields",
    );
  });

  it("rejects credential examples that encourage plaintext token handling", () => {
    const fs = makeSkill();
    fs.writeFile(
      join(SKILL_DIR, "SKILL.md"),
      '# Example\n\n```bash\necho "$API_TOKEN" | tool login\ntool login < token.txt\nexport API_TOKEN=secret\n```\n',
    );

    const report = checkCatalogFiles(SKILL_DIR, {name: "example-guide"}, fs);

    expect(report.errors).toEqual(
      expect.arrayContaining([
        "credentials: SKILL.md pipes a secret through echo",
        "credentials: SKILL.md reads a secret from token.txt",
        "credentials: SKILL.md assigns a secret in an export command",
      ]),
    );
  });

  it("allows secret exports sourced from an injected value or secret store", () => {
    const fs = makeSkill();
    fs.writeFile(
      join(SKILL_DIR, "SKILL.md"),
      '# Example\n\n```bash\nexport API_TOKEN="${INJECTED_TOKEN}"\nexport SERVICE_PAT="$(secret-store read service)"\n```\n',
    );

    const report = checkCatalogFiles(SKILL_DIR, {name: "example-guide"}, fs);

    expect(report.errors).toEqual([]);
    expect(report.passes).toContain(
      "credentials: examples avoid plaintext token anti-patterns",
    );
  });
});
