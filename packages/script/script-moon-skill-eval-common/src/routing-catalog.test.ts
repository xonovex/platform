import {mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join, resolve} from "node:path";
import {afterEach, describe, expect, it} from "vitest";
import {
  buildRoutingScenarios,
  conflictingQueryOwners,
  missingValidationRoutingOwners,
  unresolvedOperationRationales,
} from "./routing-catalog.js";

const temporaryDirectories: string[] = [];

const addSkill = (
  root: string,
  name: string,
  queries: readonly unknown[],
  references: readonly string[] = [],
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
  if (references.length === 0) return;
  mkdirSync(join(guide, "references"), {recursive: true});
  for (const reference of references) {
    writeFileSync(
      join(guide, "references", `${reference}.md`),
      `# ${reference}\n`,
    );
  }
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

  it("reports a query two skills claim, ignoring case and closing punctuation", () => {
    const root = mkdtempSync(join(tmpdir(), "routing-catalog-"));
    temporaryDirectories.push(root);
    addSkill(root, "alpha", [
      {query: "Record the decision.", should_trigger: true, split: "train"},
    ]);
    addSkill(root, "beta", [
      {query: "record the   decision", should_trigger: true, split: "train"},
    ]);
    addSkill(root, "gamma", [
      {query: "Record the decision.", should_trigger: false, split: "train"},
    ]);

    expect(conflictingQueryOwners(root)).toEqual([
      {owners: ["alpha-guide", "beta-guide"], query: "Record the decision."},
    ]);
  });

  it("leaves a query with one claimant and many negatives alone", () => {
    const root = mkdtempSync(join(tmpdir(), "routing-catalog-"));
    temporaryDirectories.push(root);
    addSkill(root, "alpha", [
      {query: "shared request", should_trigger: true, split: "train"},
    ]);
    addSkill(root, "beta", [
      {query: "shared request", should_trigger: false, split: "train"},
    ]);

    expect(conflictingQueryOwners(root)).toEqual([]);
  });

  it("reports a rationale citing an operation no skill owns a reference for", () => {
    const root = mkdtempSync(join(tmpdir(), "routing-catalog-"));
    temporaryDirectories.push(root);
    addSkill(
      root,
      "alpha",
      [
        {
          query: "settle this",
          rationale: "settling it is the alpha-decide operation",
          should_trigger: false,
          split: "train",
        },
        {
          query: "revise this",
          rationale: "this is the `alpha-revise` operation",
          should_trigger: true,
          split: "train",
        },
      ],
      ["alpha-revise"],
    );

    expect(unresolvedOperationRationales(root)).toEqual([
      {
        operation: "alpha-decide",
        rationale: "settling it is the alpha-decide operation",
        skill: "alpha-guide",
      },
    ]);
  });

  it("does not read an English compound as an operation identifier", () => {
    const root = mkdtempSync(join(tmpdir(), "routing-catalog-"));
    temporaryDirectories.push(root);
    addSkill(root, "alpha", [
      {
        query: "tag the release",
        rationale:
          "A direct version-control operation needs no workflow boundary.",
        should_trigger: false,
        split: "train",
      },
    ]);

    expect(unresolvedOperationRationales(root)).toEqual([]);
  });

  it("keeps the catalog free of claimed-twice queries and phantom operations", () => {
    const catalogRoot = resolve(import.meta.dirname, "../../../skill");

    expect(conflictingQueryOwners(catalogRoot)).toEqual([]);
    expect(unresolvedOperationRationales(catalogRoot)).toEqual([]);
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
