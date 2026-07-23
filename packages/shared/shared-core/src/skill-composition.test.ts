import {describe, expect, it} from "vitest";
import {
  compositionCatalogIssues,
  parseCompositionCatalog,
  parseCompositionRequest,
  parseInstalledSkillInventory,
  resolveComposition,
  resolveExactSkill,
  resolvePreferenceOverlays,
  resolveSemanticRequirement,
  type CompositionCatalog,
  type CompositionCatalogEntry,
  type InstalledSkill,
  type SemanticRequirement,
} from "./skill-composition.js";

const entry = (
  name: string,
  {
    functionalRole = "procedure",
    provisions = [],
    requirements = [],
  }: Partial<
    Pick<CompositionCatalogEntry, "provisions" | "requirements"> & {
      readonly functionalRole: CompositionCatalogEntry["classification"]["functionalRole"];
    }
  > = {},
): CompositionCatalogEntry => ({
  name,
  classification: {
    lifecycle: functionalRole === "preference" ? "durable" : "procedural",
    functionalRole,
  },
  provisions,
  requirements,
});

const catalog = (
  skills: readonly CompositionCatalogEntry[],
): CompositionCatalog => ({
  contractVersion: "2.0.0",
  digest: "catalog-digest",
  overlayPrecedence: [
    "global",
    "organization",
    "repository",
    "language",
    "framework",
    "path",
    "explicit",
  ],
  skills,
});

const installed = (
  guide: string,
  implementationVersion = "7.0.0",
): InstalledSkill => ({
  guide,
  implementationVersion,
  packagePath: `packages/skill/${guide.replace(/-guide$/u, "")}`,
  plugin: `xonovex-skill-${guide.replace(/-guide$/u, "")}`,
  sourcesPath: `packages/skill/${guide.replace(/-guide$/u, "")}/${guide}/SOURCES.md`,
});

const requirement = (
  overrides: Partial<SemanticRequirement> = {},
): SemanticRequirement => ({
  id: "assurance:evidence",
  range: "^1.0.0",
  strength: "required",
  reason: "The procedure needs evidence assurance.",
  ...overrides,
});

describe("composition inputs", () => {
  it("parses a v2 catalog with deterministic overlay precedence", () => {
    const result = parseCompositionCatalog({
      contractVersion: "2.0.0",
      overlayPrecedence: [
        "global",
        "organization",
        "repository",
        "language",
        "framework",
        "path",
        "explicit",
      ],
      skills: [
        {
          name: "example-guide",
          classification: {
            lifecycle: "durable",
            functionalRole: "context",
          },
        },
      ],
    });

    expect(result.success).toBe(true);
    if (!result.success) throw new Error(result.errors.join("\n"));
    expect(result.data.skills[0]).toMatchObject({
      provisions: [],
      requirements: [],
    });
    expect(result.data.digest).toMatch(/^[0-9a-f]{64}$/u);
  });

  it("rejects unsupported catalog versions and invalid precedence", () => {
    expect(
      parseCompositionCatalog({
        contractVersion: "1.0.0",
        skills: [],
      }),
    ).toEqual({
      success: false,
      errors: ["contractVersion: unsupported major 1; expected 2"],
    });
    expect(
      parseCompositionCatalog({
        contractVersion: "2.0.0",
        overlayPrecedence: Array.from({length: 7}, () => "global"),
        skills: [],
      }).success,
    ).toBe(false);
  });

  it("validates installed inventories and operation-level requests", () => {
    expect(
      parseInstalledSkillInventory([installed("example-guide")]).success,
    ).toBe(true);
    expect(
      parseCompositionRequest({
        exactSkills: [
          {
            guide: "example-guide",
            reason: "The caller selected it.",
            required: true,
          },
        ],
        requirements: [requirement({strength: "preferred"})],
        preferenceOverlays: [],
      }).success,
    ).toBe(true);
    expect(
      parseCompositionRequest({
        requirements: [
          requirement({
            range: "newest",
          }),
        ],
      }).success,
    ).toBe(false);
    expect(
      parseInstalledSkillInventory([
        installed("example-guide"),
        installed("example-guide", "8.0.0"),
      ]).success,
    ).toBe(false);
  });
});

describe("installed-snapshot selection", () => {
  it("selects an exact installed version and records provenance", () => {
    const result = resolveExactSkill(
      catalog([entry("alpha-guide")]),
      [installed("alpha-guide")],
      "alpha-guide",
      "7.0.0",
      {
        kind: "explicit",
        reason: "The caller pinned the skill.",
      },
    );

    expect(result).toEqual({
      status: "selected",
      selection: {
        catalogContractVersion: "2.0.0",
        catalogDigest: "catalog-digest",
        guide: "alpha-guide",
        implementationVersion: "7.0.0",
        packagePath: "packages/skill/alpha",
        plugin: "xonovex-skill-alpha",
        provenance: {
          kind: "explicit",
          reason: "The caller pinned the skill.",
        },
        sourcesPath: "packages/skill/alpha/alpha-guide/SOURCES.md",
      },
    });
  });

  it("returns a serializable failure when an exact skill is unavailable", () => {
    const result = resolveExactSkill(
      catalog([entry("alpha-guide")]),
      [],
      "alpha-guide",
      undefined,
      {
        kind: "exact-dependency",
        reason: "The dependency is required.",
        requestedBy: "consumer-guide",
      },
    );

    expect(result).toMatchObject({
      status: "unavailable",
      blocking: true,
      candidates: [],
      catalogContractVersion: "2.0.0",
      catalogDigest: "catalog-digest",
      requestedGuide: "alpha-guide",
    });
  });

  it("rejects duplicate installed identities instead of selecting by order", () => {
    const current = catalog([
      entry("alpha-guide", {
        provisions: [{id: "assurance:evidence", version: "1.0.0"}],
      }),
    ]);
    const inventory = [
      installed("alpha-guide", "7.0.0"),
      installed("alpha-guide", "8.0.0"),
    ];

    expect(
      resolveExactSkill(current, inventory, "alpha-guide", undefined, {
        kind: "explicit",
        reason: "The caller selected it.",
      }),
    ).toMatchObject({status: "ambiguous", blocking: true});
    expect(
      resolveSemanticRequirement(current, inventory, requirement(), {
        kind: "semantic-requirement",
        reason: "Evidence is required.",
      }),
    ).toMatchObject({status: "ambiguous", blocking: true});
  });

  it("selects one compatible semantic provider and exposes ambiguity", () => {
    const current = catalog([
      entry("alpha-guide", {
        provisions: [{id: "assurance:evidence", version: "1.2.0"}],
      }),
      entry("beta-guide", {
        provisions: [{id: "assurance:evidence", version: "2.0.0"}],
      }),
    ]);
    const selected = resolveSemanticRequirement(
      current,
      [installed("alpha-guide"), installed("beta-guide")],
      requirement(),
      {
        kind: "semantic-requirement",
        reason: "Evidence is required.",
      },
    );

    expect(selected).toMatchObject({
      status: "selected",
      selection: {
        guide: "alpha-guide",
        provision: {
          id: "assurance:evidence",
          version: "1.2.0",
        },
        requestedRange: "^1.0.0",
      },
    });

    const ambiguous = resolveSemanticRequirement(
      catalog([
        entry("alpha-guide", {
          provisions: [{id: "assurance:evidence", version: "1.0.0"}],
        }),
        entry("beta-guide", {
          provisions: [{id: "assurance:evidence", version: "1.1.0"}],
        }),
      ]),
      [installed("alpha-guide"), installed("beta-guide")],
      requirement(),
      {
        kind: "policy",
        reason: "Policy requires evidence.",
      },
    );
    expect(ambiguous).toMatchObject({
      status: "ambiguous",
      candidates: ["alpha-guide@1.0.0", "beta-guide@1.1.0"],
      requirement: {
        id: "assurance:evidence",
      },
    });
  });

  it("preserves unresolved preferred requirements as visible issues", () => {
    const current = catalog([
      entry("alpha-guide", {
        requirements: [requirement({strength: "preferred"})],
      }),
    ]);

    const issues = compositionCatalogIssues(current, [
      installed("alpha-guide"),
    ]);

    expect(issues.errors).toEqual([]);
    expect(issues.preferredFailures).toMatchObject([
      {
        status: "unavailable",
        blocking: false,
        requirement: {
          id: "assurance:evidence",
          strength: "preferred",
        },
      },
    ]);
  });
});

describe("preference overlays", () => {
  it("orders applicable overlays from broad to narrow scope", () => {
    const current = catalog([
      entry("base-preference-guide", {functionalRole: "preference"}),
      entry("local-preference-guide", {functionalRole: "preference"}),
    ]);

    const resolution = resolvePreferenceOverlays(
      current,
      [installed("base-preference-guide"), installed("local-preference-guide")],
      [
        {
          guide: "local-preference-guide",
          target: "language:c99",
          scope: {kind: "repository", value: "example/repo"},
          reason: "The repository selects local conventions.",
        },
        {
          guide: "base-preference-guide",
          target: "language:c99",
          scope: {kind: "global", value: "*"},
          reason: "The global defaults apply.",
        },
      ],
    );

    expect(resolution.conflicts).toEqual([]);
    expect(resolution.applied.map(({guide}) => guide)).toEqual([
      "base-preference-guide",
      "local-preference-guide",
    ]);
  });

  it("rejects fact skills and equal-scope preference conflicts", () => {
    const current = catalog([
      entry("domain-guide", {functionalRole: "domain"}),
      entry("first-preference-guide", {functionalRole: "preference"}),
      entry("second-preference-guide", {functionalRole: "preference"}),
    ]);
    const scope = {kind: "repository" as const, value: "example/repo"};

    const resolution = resolvePreferenceOverlays(
      current,
      [
        installed("domain-guide"),
        installed("first-preference-guide"),
        installed("second-preference-guide"),
      ],
      [
        {
          guide: "domain-guide",
          target: "domain:orders",
          scope,
          reason: "Invalid factual override.",
        },
        {
          guide: "first-preference-guide",
          target: "language:c99",
          scope,
          reason: "First convention.",
        },
        {
          guide: "second-preference-guide",
          target: "language:c99",
          scope,
          reason: "Second convention.",
        },
      ],
    );

    expect(resolution.failures).toMatchObject([
      {
        status: "conflict",
        candidates: ["domain-guide"],
      },
    ]);
    expect(resolution.conflicts).toMatchObject([
      {
        status: "conflict",
        candidates: ["first-preference-guide", "second-preference-guide"],
      },
    ]);
    expect(resolution.applied).toEqual([]);
  });
});

describe("composition runtime", () => {
  it("resolves invariant and requested requirements without hiding degradation", () => {
    const current = catalog([
      entry("assurance-guide", {
        functionalRole: "assurance",
        provisions: [{id: "assurance:evidence", version: "1.0.0"}],
      }),
      entry("procedure-guide", {
        requirements: [requirement()],
      }),
    ]);

    const resolution = resolveComposition(
      current,
      [installed("assurance-guide"), installed("procedure-guide")],
      {
        exactSkills: [
          {
            guide: "procedure-guide",
            reason: "The operation selected its owner.",
            required: true,
          },
        ],
        requirements: [
          requirement({
            id: "assurance:optional",
            strength: "preferred",
            reason: "Optional specialist review.",
          }),
        ],
        preferenceOverlays: [],
      },
    );

    expect(resolution.status).toBe("degraded");
    expect(resolution.skills).toMatchObject([
      {
        status: "selected",
        guide: "procedure-guide",
      },
      {
        status: "selected",
        guide: "assurance-guide",
      },
      {
        status: "unavailable",
        required: false,
        blocking: false,
      },
    ]);
  });

  it("resolves transitive invariant requirements from semantic providers and overlays", () => {
    const current = catalog([
      entry("base-preference-guide", {
        functionalRole: "preference",
        requirements: [
          requirement({
            id: "context:policy",
            reason: "The preference needs policy context.",
          }),
        ],
      }),
      entry("context-guide", {
        functionalRole: "context",
        provisions: [{id: "context:policy", version: "1.0.0"}],
      }),
      entry("procedure-guide", {
        provisions: [{id: "procedure:work", version: "1.0.0"}],
        requirements: [requirement()],
      }),
      entry("assurance-guide", {
        functionalRole: "assurance",
        provisions: [{id: "assurance:evidence", version: "1.0.0"}],
        requirements: [
          requirement({
            id: "context:policy",
            reason: "The assurance needs policy context.",
          }),
        ],
      }),
    ]);

    const resolution = resolveComposition(
      current,
      [
        installed("assurance-guide"),
        installed("base-preference-guide"),
        installed("context-guide"),
        installed("procedure-guide"),
      ],
      {
        exactSkills: [],
        requirements: [
          requirement({
            id: "procedure:work",
            reason: "The operation needs its procedure.",
          }),
        ],
        preferenceOverlays: [
          {
            guide: "base-preference-guide",
            target: "language:c99",
            scope: {kind: "global", value: "*"},
            reason: "Apply global C conventions.",
          },
        ],
      },
    );

    expect(resolution.status).toBe("ready");
    expect(
      resolution.skills.map((selection) =>
        selection.status === "selected" ? selection.guide : selection.status,
      ),
    ).toEqual([
      "procedure-guide",
      "assurance-guide",
      "context-guide",
      "context-guide",
    ]);
  });

  it("does not resolve dependencies of an exact skill that was not selected", () => {
    const current = catalog([
      entry("procedure-guide", {
        requirements: [requirement()],
      }),
    ]);

    const resolution = resolveComposition(current, [], {
      exactSkills: [
        {
          guide: "procedure-guide",
          reason: "The operation needs its procedure.",
          required: true,
        },
      ],
      requirements: [],
      preferenceOverlays: [],
    });

    expect(resolution.skills).toHaveLength(1);
    expect(resolution.status).toBe("blocked");
  });
});
