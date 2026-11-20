import {join} from "node:path";
import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {describe, expect, it} from "vitest";
import {
  buildRoutingScenarios,
  catalogQueryOwners,
  conflictingQueryOwners,
  missingValidationRoutingOwners,
  unresolvedOperationRationales,
} from "../../../src/routing-catalog.js";

const ROOT = "/catalog";

interface Skill {
  readonly name: string;
  readonly plugin?: string;
  readonly queries: readonly unknown[];
  readonly references?: readonly string[];
}

// A catalog is a set of grouped plugins, each holding skills with their queries
// and, optionally, the reference files an operation rationale can name.
const catalog = (...skills: readonly Skill[]): FileSystem => {
  const files: Record<string, string> = {};
  for (const {name, plugin: group = name, queries, references = []} of skills) {
    const plugin = join(ROOT, `plugin-${group}`);
    const skill = join(plugin, "skills", `using-${name}`);
    files[join(plugin, ".claude-plugin", "plugin.json")] = JSON.stringify({
      name: `xonovex-${group}`,
      skills: "./skills/",
    });
    files[join(skill, "SKILL.md")] =
      `---\nname: using-${name}\ndescription: Use for ${name}.\n---\n`;
    files[join(skill, "eval-queries.json")] = JSON.stringify(queries);
    for (const reference of references) {
      files[join(skill, "references", `${reference}.md`)] = `# ${reference}\n`;
    }
  }
  return memoryFileSystem({files});
};

describe("routing catalog", () => {
  it("keeps skill ownership distinct inside one grouped plugin", () => {
    const fs = catalog(
      {
        name: "alpha",
        plugin: "core",
        queries: [
          {query: "alpha request", should_trigger: true, split: "validation"},
          {query: "beta request", should_trigger: false, split: "validation"},
          {query: "ambiguous request", should_trigger: true},
          {query: "unclaimed request", should_trigger: false},
        ],
      },
      {
        name: "beta",
        plugin: "core",
        queries: [
          {query: "alpha request", should_trigger: false, split: "validation"},
          {query: "beta request", should_trigger: true, split: "validation"},
          {query: "ambiguous request", should_trigger: true},
        ],
      },
    );

    expect([...catalogQueryOwners(ROOT, fs)]).toEqual([
      ["alpha request", "using-alpha"],
      ["beta request", "using-beta"],
    ]);
    const scenarios = buildRoutingScenarios(ROOT, fs);
    expect(scenarios.map(({expectedSkill}) => expectedSkill)).toEqual([
      "using-alpha",
      "using-beta",
    ]);
    expect(
      scenarios[0]?.candidates.map(({pluginDirectory}) => pluginDirectory),
    ).toEqual([join(ROOT, "plugin-core"), join(ROOT, "plugin-core")]);
  });

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
        expectedSkill: "using-alpha",
        candidates: [
          expect.objectContaining({shortName: "using-alpha"}),
          expect.objectContaining({shortName: "using-beta"}),
        ],
      },
    ]);
    expect(missingValidationRoutingOwners(ROOT, fs)).toEqual(["using-beta"]);
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
      {owners: ["using-alpha", "using-beta"], query: "Record the decision."},
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
          rationale: "settling it is the `alpha-decide` operation",
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
        rationale: "settling it is the `alpha-decide` operation",
        skill: "using-alpha",
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
