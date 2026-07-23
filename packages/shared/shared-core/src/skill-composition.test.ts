import {describe, expect, it} from "vitest";
import {
  parseCompositionCatalog,
  parseCompositionRequest,
  parseInstalledSkillInventory,
  type CompositionCatalog,
  type CompositionCatalogEntry,
  type CompositionRequest,
  type InstalledSkill,
  type InstalledSkillDependency,
  type SemanticRequirement,
} from "./skill-composition-contract.js";
import {compositionCatalogIssues} from "./skill-composition-validation.js";
import {resolveComposition} from "./skill-composition.js";
import {resolvePreferenceOverlays} from "./skill-preference-overlays.js";
import {
  resolveExactSkill,
  resolveSemanticRequirement,
} from "./skill-selection.js";

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
  dependencies: readonly InstalledSkillDependency[] = [],
): InstalledSkill => ({
  guide,
  implementationVersion,
  dependencies: [...dependencies],
  packagePath: `packages/skill/${guide.replace(/-guide$/u, "")}`,
  plugin: `xonovex-skill-${guide.replace(/-guide$/u, "")}`,
  sourcesPath: `packages/skill/${guide.replace(/-guide$/u, "")}/${guide}/SOURCES.md`,
});

const compositionRequest = (
  overrides: Partial<CompositionRequest>,
): CompositionRequest => ({
  exactSkills: [],
  overlayContext: {
    global: "*",
    languages: [],
    frameworks: [],
    paths: [],
    explicit: [],
  },
  preferenceOverlays: [],
  requirementProviders: {},
  requirements: [],
  ...overrides,
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
    const input = {
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
    };
    const result = parseCompositionCatalog(input, JSON.stringify(input));

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
      parseCompositionCatalog(
        {
          contractVersion: "1.0.0",
          skills: [],
        },
        '{"contractVersion":"1.0.0","skills":[]}',
      ),
    ).toEqual({
      success: false,
      errors: ["contractVersion: unsupported major 1; expected 2"],
    });
    expect(
      parseCompositionCatalog(
        {
          contractVersion: "2.0.0",
          overlayPrecedence: Array.from({length: 7}, () => "global"),
          skills: [],
        },
        '{"contractVersion":"2.0.0","skills":[]}',
      ).success,
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
      {
        global: "*",
        repository: "example/repo",
        languages: [],
        frameworks: [],
        paths: [],
        explicit: [],
      },
    );

    expect(resolution.conflicts).toEqual([]);
    expect(resolution.applied.map(({guide}) => guide)).toEqual([
      "local-preference-guide",
    ]);
    expect(resolution.shadowed.map(({guide}) => guide)).toEqual([
      "base-preference-guide",
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
      {
        global: "*",
        repository: "example/repo",
        languages: [],
        frameworks: [],
        paths: [],
        explicit: [],
      },
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

  it("rejects incomparable applicable overlays and skips unrelated scopes", () => {
    const current = catalog([
      entry("language-preference-guide", {functionalRole: "preference"}),
      entry("repository-preference-guide", {functionalRole: "preference"}),
      entry("unused-preference-guide", {functionalRole: "preference"}),
    ]);

    const resolution = resolvePreferenceOverlays(
      current,
      [
        installed("language-preference-guide"),
        installed("repository-preference-guide"),
        installed("unused-preference-guide"),
      ],
      [
        {
          guide: "repository-preference-guide",
          target: "formatting:source",
          scope: {kind: "repository", value: "example/repo"},
          reason: "The repository has formatting conventions.",
        },
        {
          guide: "language-preference-guide",
          target: "formatting:source",
          scope: {kind: "language", value: "typescript"},
          reason: "The language has formatting conventions.",
        },
        {
          guide: "unused-preference-guide",
          target: "formatting:source",
          scope: {kind: "framework", value: "react"},
          reason: "The framework convention is unrelated.",
        },
      ],
      {
        global: "*",
        repository: "example/repo",
        languages: ["typescript"],
        frameworks: [],
        paths: [],
        explicit: [],
      },
    );

    expect(resolution.applied).toEqual([]);
    expect(resolution.conflicts[0]).toMatchObject({
      candidates: ["language-preference-guide", "repository-preference-guide"],
    });
    expect(resolution.skipped.map(({guide}) => guide)).toEqual([
      "unused-preference-guide",
    ]);
  });

  it("lets an explicit overlay dominate otherwise incomparable scopes", () => {
    const current = catalog([
      entry("language-preference-guide", {functionalRole: "preference"}),
      entry("repository-preference-guide", {functionalRole: "preference"}),
      entry("explicit-preference-guide", {functionalRole: "preference"}),
    ]);

    const resolution = resolvePreferenceOverlays(
      current,
      [
        installed("explicit-preference-guide"),
        installed("language-preference-guide"),
        installed("repository-preference-guide"),
      ],
      [
        {
          guide: "repository-preference-guide",
          target: "formatting:source",
          scope: {kind: "repository", value: "example/repo"},
          reason: "The repository has formatting conventions.",
        },
        {
          guide: "language-preference-guide",
          target: "formatting:source",
          scope: {kind: "language", value: "typescript"},
          reason: "The language has formatting conventions.",
        },
        {
          guide: "explicit-preference-guide",
          target: "formatting:source",
          scope: {kind: "explicit", value: "release"},
          reason: "The release explicitly selects its convention.",
        },
      ],
      {
        global: "*",
        repository: "example/repo",
        languages: ["typescript"],
        frameworks: [],
        paths: [],
        explicit: ["release"],
      },
    );

    expect(resolution.conflicts).toEqual([]);
    expect(resolution.applied.map(({guide}) => guide)).toEqual([
      "explicit-preference-guide",
    ]);
    expect(resolution.shadowed.map(({guide}) => guide)).toEqual([
      "repository-preference-guide",
      "language-preference-guide",
    ]);
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
      compositionRequest({
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
      }),
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
      compositionRequest({
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
      }),
    );

    expect(resolution.status).toBe("ready");
    expect(
      resolution.skills.map((selection) =>
        selection.status === "selected" ? selection.guide : selection.status,
      ),
    ).toEqual([
      "procedure-guide",
      "base-preference-guide",
      "assurance-guide",
      "context-guide",
    ]);
  });

  it("does not resolve dependencies of an exact skill that was not selected", () => {
    const current = catalog([
      entry("procedure-guide", {
        requirements: [requirement()],
      }),
    ]);

    const resolution = resolveComposition(
      current,
      [],
      compositionRequest({
        exactSkills: [
          {
            guide: "procedure-guide",
            reason: "The operation needs its procedure.",
            required: true,
          },
        ],
      }),
    );

    expect(resolution.skills).toHaveLength(1);
    expect(resolution.status).toBe("blocked");
  });

  it("uses an exact selection to disambiguate and aggregates duplicate edges", () => {
    const current = catalog([
      entry("first-guide", {
        provisions: [{id: "assurance:evidence", version: "1.0.0"}],
      }),
      entry("second-guide", {
        provisions: [{id: "assurance:evidence", version: "1.1.0"}],
      }),
    ]);

    const resolution = resolveComposition(
      current,
      [installed("first-guide"), installed("second-guide")],
      compositionRequest({
        exactSkills: [
          {
            guide: "first-guide",
            reason: "The caller selected the first provider.",
            required: true,
          },
        ],
        requirements: [requirement()],
      }),
    );

    expect(resolution.status).toBe("ready");
    expect(resolution.loadOrder).toEqual(["first-guide"]);
    expect(resolution.skills).toHaveLength(1);
    expect(resolution.skills[0]).toMatchObject({
      guide: "first-guide",
      reasons: [
        "The caller selected the first provider.",
        "The procedure needs evidence assurance.",
      ],
      provenances: [{kind: "explicit"}, {kind: "semantic-requirement"}],
    });
  });

  it("loads exact hard dependencies before their consumer", () => {
    const current = catalog([entry("base-guide"), entry("consumer-guide")]);

    const resolution = resolveComposition(
      current,
      [
        installed("base-guide"),
        installed("consumer-guide", "7.0.0", [
          {
            plugin: "xonovex-skill-base",
            implementationVersion: "7.0.0",
          },
        ]),
      ],
      compositionRequest({
        exactSkills: [
          {
            guide: "consumer-guide",
            reason: "The operation needs the consumer.",
            required: true,
          },
        ],
      }),
    );

    expect(resolution.status).toBe("ready");
    expect(resolution.loadOrder).toEqual(["base-guide", "consumer-guide"]);
    expect(resolution.skills[0]).toMatchObject({
      guide: "base-guide",
      provenance: {
        kind: "exact-dependency",
        requestedBy: "consumer-guide",
      },
    });
  });

  it("closes required invariants before resolving preferred support", () => {
    const current = catalog([
      entry("owner-guide", {
        requirements: [
          requirement({
            id: "capability:base",
            reason: "The owner requires its base provider.",
          }),
        ],
      }),
      entry("base-guide", {
        provisions: [
          {id: "capability:base", version: "1.0.0"},
          {id: "assurance:evidence", version: "1.0.0"},
        ],
      }),
      entry("alternative-guide", {
        provisions: [{id: "assurance:evidence", version: "1.1.0"}],
      }),
    ]);

    const resolution = resolveComposition(
      current,
      [
        installed("alternative-guide"),
        installed("base-guide"),
        installed("owner-guide"),
      ],
      compositionRequest({
        exactSkills: [
          {
            guide: "owner-guide",
            reason: "The operation selects its owner.",
            required: true,
          },
        ],
        requirements: [requirement({strength: "preferred"})],
      }),
    );

    expect(resolution.status).toBe("ready");
    expect(resolution.loadOrder).toEqual(["owner-guide", "base-guide"]);
    expect(resolution.skills).toHaveLength(2);
    expect(resolution.skills[1]).toMatchObject({
      guide: "base-guide",
      matchedProvisions: [
        {provision: {id: "capability:base"}},
        {provision: {id: "assurance:evidence"}},
      ],
    });
  });
});
