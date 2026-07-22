import {
  mkdirSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import {tmpdir} from "node:os";
import {join, resolve} from "node:path";
import {afterEach, describe, expect, it} from "vitest";
import {catalogStats, renderCatalogStats} from "./catalog-stats.mjs";

const workspaceRoot = resolve(import.meta.dirname, "../../../../..");
const temporaryDirectories = [];

const addSkill = (root, name, tier, shouldTrigger) => {
  const guide = join(root, `skill-${name}`, `${name}-guide`);
  mkdirSync(join(guide, "references"), {recursive: true});
  writeFileSync(
    join(guide, "SKILL.md"),
    `---\nname: ${name}-guide\ndescription: "Use when testing ${name}."\n---\n# ${name}\n`,
  );
  writeFileSync(join(guide, "references", "details.md"), "# Details\n");
  writeFileSync(
    join(guide, "evals.json"),
    JSON.stringify({tier, evals: [{prompt: "one"}, {prompt: "two"}]}),
  );
  writeFileSync(
    join(guide, "eval-queries.json"),
    JSON.stringify([
      {
        query: "shared validation request",
        should_trigger: shouldTrigger,
        split: "validation",
      },
    ]),
  );
};

afterEach(() => {
  for (const directory of temporaryDirectories) {
    rmSync(directory, {recursive: true, force: true});
  }
  temporaryDirectories.length = 0;
});

describe("catalogStats", () => {
  it("counts catalog evidence and competitive routing", () => {
    const root = mkdtempSync(join(tmpdir(), "catalog-stats-"));
    temporaryDirectories.push(root);
    addSkill(root, "alpha", "aggressive", true);
    addSkill(root, "beta", "moderate", false);

    const stats = catalogStats(root);

    expect(stats).toEqual({
      distilledLines: 12,
      outputEvals: 4,
      routing: {
        scenarios: 1,
        train: 0,
        validation: 1,
        validationOwners: 1,
      },
      skills: 2,
      tiers: {aggressive: 1, moderate: 1, conservative: 0},
      triggerQueries: 2,
    });
  });

  it("keeps the optimization audit synchronized with the catalog", () => {
    const catalogRoot = join(workspaceRoot, "packages/skill");
    const audit = readFileSync(
      join(catalogRoot, "optimization-audit.md"),
      "utf8",
    );
    const generated =
      /<!-- catalog-stats:start -->\n([\s\S]*?)\n<!-- catalog-stats:end -->/
        .exec(audit)?.[1]
        ?.trim();

    const expected = renderCatalogStats(catalogStats(catalogRoot));

    expect(generated).toBe(expected);
  });
});
