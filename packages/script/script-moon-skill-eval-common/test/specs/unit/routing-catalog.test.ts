import {join} from "node:path";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {memoryFileSystem} from "@xonovex/script-moon-common/file-system-memory";
import {describe, expect, it} from "vitest";
import {
  buildRoutingScenarios,
  conflictingQueryOwners,
  missingValidationRoutingOwners,
  unresolvedOperationRationales,
} from "../../../src/routing-catalog.js";

const ROOT = "/catalog";

interface Skill {
  readonly name: string;
  readonly queries: readonly unknown[];
  readonly references?: readonly string[];
}

// A catalog is a set of skill plugins, each holding one guide with its queries
// and, optionally, the reference files an operation rationale can name.
const catalog = (...skills: readonly Skill[]): FileSystem => {
  const files: Record<string, string> = {};
  for (const {name, queries, references = []} of skills) {
    const plugin = join(ROOT, `skill-${name}`);
    const guide = join(plugin, `${name}-guide`);
    files[join(plugin, ".claude-plugin", "plugin.json")] = JSON.stringify({
      name: `xonovex-skill-${name}`,
      skills: [`./${name}-guide`],
    });
    files[join(guide, "SKILL.md")] =
      `---\nname: ${name}-guide\ndescription: Use for ${name}.\n---\n`;
    files[join(guide, "eval-queries.json")] = JSON.stringify(queries);
    for (const reference of references) {
      files[join(guide, "references", `${reference}.md`)] = `# ${reference}\n`;
    }
  }
  return memoryFileSystem({files});
};

describe("routing catalog", () => {
  it("builds competitive scenarios from one positive and shared negatives", () => {
    const fs = catalog(
      {
        name: "alpha",
        queries: [
          {query: "shared request", should_trigger: true, split: "validation"},
          {query: "alpha only", should_trigger: true, split: "train"},
        ],
      },
      {
        name: "beta",
        queries: [
          {query: "shared request", should_trigger: false, split: "validation"},
        ],
      },
    );

    expect(buildRoutingScenarios(ROOT, fs)).toEqual([
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
    expect(missingValidationRoutingOwners(ROOT, fs)).toEqual(["beta-guide"]);
  });

  it("reports a query two skills claim, ignoring case and closing punctuation", () => {
    const fs = catalog(
      {
        name: "alpha",
        queries: [
          {query: "Record the decision.", should_trigger: true, split: "train"},
        ],
      },
      {
        name: "beta",
        queries: [
          {
            query: "record the   decision",
            should_trigger: true,
            split: "train",
          },
        ],
      },
      {
        name: "gamma",
        queries: [
          {
            query: "Record the decision.",
            should_trigger: false,
            split: "train",
          },
        ],
      },
    );

    expect(conflictingQueryOwners(ROOT, fs)).toEqual([
      {owners: ["alpha-guide", "beta-guide"], query: "Record the decision."},
    ]);
  });

  it("leaves a query with one claimant and many negatives alone", () => {
    const fs = catalog(
      {
        name: "alpha",
        queries: [
          {query: "shared request", should_trigger: true, split: "train"},
        ],
      },
      {
        name: "beta",
        queries: [
          {query: "shared request", should_trigger: false, split: "train"},
        ],
      },
    );

    expect(conflictingQueryOwners(ROOT, fs)).toEqual([]);
  });

  it("reports a rationale citing an operation no skill owns a reference for", () => {
    const fs = catalog({
      name: "alpha",
      references: ["alpha-revise"],
      queries: [
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
    });

    expect(unresolvedOperationRationales(ROOT, fs)).toEqual([
      {
        operation: "alpha-decide",
        rationale: "settling it is the alpha-decide operation",
        skill: "alpha-guide",
      },
    ]);
  });

  it("does not read an English compound as an operation identifier", () => {
    const fs = catalog({
      name: "alpha",
      queries: [
        {
          query: "tag the release",
          rationale:
            "A direct version-control operation needs no workflow boundary.",
          should_trigger: false,
          split: "train",
        },
      ],
    });

    expect(unresolvedOperationRationales(ROOT, fs)).toEqual([]);
  });

  it("rejects missing roots and malformed query files", () => {
    expect(() =>
      buildRoutingScenarios("/missing-routing-catalog", memoryFileSystem()),
    ).toThrow("catalog root not found");

    expect(() =>
      buildRoutingScenarios(ROOT, catalog({name: "alpha", queries: []})),
    ).toThrow("invalid queries");
  });
});
