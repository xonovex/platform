import {mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join, resolve} from "node:path";
import {afterEach, describe, expect, it} from "vitest";
import {
  buildRoutingScenarios,
  missingValidationRoutingOwners,
} from "./routing-catalog.js";

const temporaryDirectories: string[] = [];

const addSkill = (
  root: string,
  name: string,
  queries: readonly unknown[],
): void => {
  const plugin = join(root, `skill-${name}`);
  const guide = join(plugin, `${name}-guide`);
  mkdirSync(join(plugin, ".claude-plugin"), {recursive: true});
  mkdirSync(guide, {recursive: true});
  writeFileSync(
    join(plugin, ".claude-plugin", "plugin.json"),
    JSON.stringify({
      name: `xonovex-skill-${name}`,
      skills: [`./${name}-guide`],
    }),
  );
  writeFileSync(
    join(guide, "SKILL.md"),
    `---\nname: ${name}-guide\ndescription: Use for ${name}.\n---\n`,
  );
  writeFileSync(join(guide, "eval-queries.json"), JSON.stringify(queries));
};

afterEach(() => {
  for (const directory of temporaryDirectories) {
    rmSync(directory, {recursive: true, force: true});
  }
  temporaryDirectories.length = 0;
});

describe("routing catalog", () => {
  it("builds competitive scenarios from one positive and shared negatives", () => {
    const root = mkdtempSync(join(tmpdir(), "routing-catalog-"));
    temporaryDirectories.push(root);
    addSkill(root, "alpha", [
      {query: "shared request", should_trigger: true, split: "validation"},
      {query: "alpha only", should_trigger: true, split: "train"},
    ]);
    addSkill(root, "beta", [
      {query: "shared request", should_trigger: false, split: "validation"},
    ]);

    expect(buildRoutingScenarios(root)).toEqual([
      {
        query: "shared request",
        split: "validation",
        expectedSkill: "alpha-guide",
        candidates: [
          expect.objectContaining({shortName: "alpha-guide"}),
          expect.objectContaining({shortName: "beta-guide"}),
        ],
      },
    ]);
    expect(missingValidationRoutingOwners(root)).toEqual(["beta-guide"]);
  });

  it("gives every catalog skill a validation routing scenario", () => {
    const catalogRoot = resolve(import.meta.dirname, "../../../skill");

    const missingOwners = missingValidationRoutingOwners(catalogRoot);

    expect(missingOwners).toEqual([]);
  });

  it("rejects missing roots and malformed query files", () => {
    expect(() => buildRoutingScenarios("/missing-routing-catalog")).toThrow(
      "catalog root not found",
    );
    const root = mkdtempSync(join(tmpdir(), "routing-catalog-"));
    temporaryDirectories.push(root);
    addSkill(root, "alpha", []);
    expect(() => buildRoutingScenarios(root)).toThrow("invalid queries");
  });
});
