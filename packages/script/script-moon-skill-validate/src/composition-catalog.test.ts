import {readFileSync} from "node:fs";
import {resolve} from "node:path";
import {
  parseCompositionCatalog,
  type CompositionCatalog,
  type CompositionCatalogEntry,
  type InstalledSkill,
  type SemanticRequirement,
} from "@xonovex/core/skill-composition-contract";
import {
  compositionCatalogIssues,
  compositionCatalogSnapshotErrors,
} from "@xonovex/core/skill-composition-validation";
import {
  resolveExactSkill,
  resolveSemanticRequirement,
} from "@xonovex/core/skill-selection";
import {describe, expect, it} from "vitest";
import {checkCompositionCatalog} from "./composition-catalog-check.js";
import {type LinkReport} from "./reference-file-links.js";

const compositionCatalogErrors = (
  currentCatalog: CompositionCatalog,
  installedSkills: readonly InstalledSkill[],
): readonly string[] =>
  compositionCatalogIssues(currentCatalog, installedSkills).errors;

const entry = (
  name: string,
  {
    provisions = [],
    requirements = [],
  }: Partial<Pick<CompositionCatalogEntry, "provisions" | "requirements">> = {},
): CompositionCatalogEntry => ({
  name,
  classification: {
    lifecycle: "procedural",
    functionalRole: "procedure",
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
  dependencies: [],
  packagePath: `packages/skill/${guide.replace(/-guide$/, "")}`,
  plugin: `xonovex-skill-${guide.replace(/-guide$/, "")}`,
  sourcesPath: `packages/skill/${guide.replace(/-guide$/, "")}/${guide}/SOURCES.md`,
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

const parseCatalog = (input: unknown) =>
  parseCompositionCatalog(input, JSON.stringify(input));

describe("composition catalog schema", () => {
  it("accepts independent lifecycle and functional-role classifications", () => {
    const result = parseCatalog({
      contractVersion: "2.0.0",
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
    expect(result.data.digest).toMatch(/^[0-9a-f]{64}$/);
  });

  it.each([
    [
      "mixed lifecycle",
      {
        contractVersion: "2.0.0",
        skills: [
          {
            name: "example-guide",
            classification: {
              lifecycle: "mixed",
              functionalRole: "procedure",
            },
          },
        ],
      },
    ],
    [
      "mixed functional role",
      {
        contractVersion: "2.0.0",
        skills: [
          {
            name: "example-guide",
            classification: {
              lifecycle: "procedural",
              functionalRole: "mixed",
            },
          },
        ],
      },
    ],
    [
      "invalid provision version",
      {
        contractVersion: "2.0.0",
        skills: [
          {
            name: "example-guide",
            classification: {
              lifecycle: "procedural",
              functionalRole: "assurance",
            },
            provisions: [{id: "assurance:evidence", version: "latest"}],
          },
        ],
      },
    ],
    [
      "invalid requirement range",
      {
        contractVersion: "2.0.0",
        skills: [
          {
            name: "example-guide",
            classification: {
              lifecycle: "procedural",
              functionalRole: "procedure",
            },
            requirements: [
              {
                id: "assurance:evidence",
                range: "newest",
                strength: "required",
                reason: "Evidence is mandatory.",
              },
            ],
          },
        ],
      },
    ],
  ])("rejects %s", (_label, input) => {
    expect(parseCatalog(input).success).toBe(false);
  });

  it("rejects an unsupported catalog contract major", () => {
    const result = parseCatalog({
      contractVersion: "3.0.0",
      skills: [],
    });

    expect(result).toEqual({
      success: false,
      errors: ["contractVersion: unsupported major 3; expected 2"],
    });
  });

  it("requires the packaged workflow snapshot to be byte-identical", () => {
    expect(
      compositionCatalogSnapshotErrors('{"skills":[]}\n', undefined),
    ).toEqual([
      "packaged workflow snapshot workflow-guide/assets/composition-catalog.json is missing",
    ]);
    expect(
      compositionCatalogSnapshotErrors('{"skills":[]}\n', '{"skills": []}\n'),
    ).toEqual([
      "packaged workflow snapshot differs from packages/skill/composition-catalog.json",
    ]);
    expect(
      compositionCatalogSnapshotErrors('{"skills":[]}\n', '{"skills":[]}\n'),
    ).toEqual([]);
  });

  it("validates the complete repository catalog", () => {
    const repositoryRoot = resolve(import.meta.dirname, "../../../..");
    const passes: string[] = [];
    const failures: string[] = [];
    const report: LinkReport = {
      addPass: (message) => {
        passes.push(message);
      },
      addFail: (message) => {
        failures.push(message);
      },
    };

    checkCompositionCatalog(repositoryRoot, report);

    expect(failures).toEqual([]);
    expect(passes).toContain(
      "composition catalog: 92 installed skill(s) have one validated classification",
    );
    const parsed = JSON.parse(
      readFileSync(
        resolve(repositoryRoot, "packages/skill/composition-catalog.json"),
        "utf8",
      ),
    ) as {skills: unknown[]};
    expect(parsed.skills).toHaveLength(92);
  });
});

describe("composition catalog integrity", () => {
  it("reports duplicate, unsorted, missing, and unexpected entries", () => {
    const errors = compositionCatalogErrors(
      catalog([
        entry("zulu-guide"),
        entry("alpha-guide"),
        entry("alpha-guide"),
      ]),
      [installed("alpha-guide"), installed("missing-guide")],
    );

    expect(errors).toEqual(
      expect.arrayContaining([
        "duplicate skill entry alpha-guide",
        "skill entries must be sorted alphabetically by name",
        "missing installed skill missing-guide",
        "catalog skill zulu-guide is not installed",
      ]),
    );
  });

  it("rejects duplicate provisions and requirements within one skill", () => {
    const repeatedRequirement = requirement();
    const errors = compositionCatalogErrors(
      catalog([
        entry("alpha-guide", {
          provisions: [
            {id: "assurance:evidence", version: "1.0.0"},
            {id: "assurance:evidence", version: "2.0.0"},
          ],
          requirements: [repeatedRequirement, repeatedRequirement],
        }),
      ]),
      [installed("alpha-guide")],
    );

    expect(errors).toEqual(
      expect.arrayContaining([
        "alpha-guide declares provision assurance:evidence more than once",
        "alpha-guide declares requirement assurance:evidence more than once",
      ]),
    );
  });

  it("rejects duplicate requirement identifiers and installed guide identities", () => {
    const errors = compositionCatalogErrors(
      catalog([
        entry("alpha-guide", {
          requirements: [
            requirement(),
            requirement({range: "^2.0.0", strength: "preferred"}),
          ],
        }),
      ]),
      [installed("alpha-guide"), installed("alpha-guide", "8.0.0")],
    );

    expect(errors).toEqual(
      expect.arrayContaining([
        "alpha-guide declares requirement assurance:evidence more than once",
        "installed snapshot contains duplicate guide alpha-guide",
      ]),
    );
  });

  it("accepts an unresolved preferred requirement without hiding it", () => {
    const errors = compositionCatalogErrors(
      catalog([
        entry("alpha-guide", {
          requirements: [requirement({strength: "preferred"})],
        }),
      ]),
      [installed("alpha-guide")],
    );

    expect(errors).toEqual([]);
    const resolution = resolveSemanticRequirement(
      catalog([entry("alpha-guide")]),
      [installed("alpha-guide")],
      requirement({strength: "preferred"}),
      {
        kind: "semantic-requirement",
        reason: "Preferred evidence support.",
        requestedBy: "alpha-guide",
      },
    );
    expect(resolution).toMatchObject({
      status: "unavailable",
      blocking: false,
      candidates: [],
    });
  });

  it("rejects unavailable and ambiguous required requirements", () => {
    const unavailableCatalog = catalog([
      entry("alpha-guide", {requirements: [requirement()]}),
    ]);
    expect(
      compositionCatalogErrors(unavailableCatalog, [installed("alpha-guide")]),
    ).toContain(
      "alpha-guide required requirement is not deterministic: assurance:evidence has no installed provider",
    );

    const ambiguousCatalog = catalog([
      entry("alpha-guide", {requirements: [requirement()]}),
      entry("beta-guide", {
        provisions: [{id: "assurance:evidence", version: "1.0.0"}],
      }),
      entry("gamma-guide", {
        provisions: [{id: "assurance:evidence", version: "1.1.0"}],
      }),
    ]);
    expect(
      compositionCatalogErrors(ambiguousCatalog, [
        installed("alpha-guide"),
        installed("beta-guide"),
        installed("gamma-guide"),
      ]),
    ).toContain(
      "alpha-guide required requirement is not deterministic: assurance:evidence@^1.0.0 has multiple compatible installed providers",
    );
  });

  it("rejects cycles among deterministically resolved required requirements", () => {
    const cyclic = catalog([
      entry("alpha-guide", {
        provisions: [{id: "procedure:alpha", version: "1.0.0"}],
        requirements: [
          requirement({
            id: "procedure:beta",
            reason: "Alpha requires beta.",
          }),
        ],
      }),
      entry("beta-guide", {
        provisions: [{id: "procedure:beta", version: "1.0.0"}],
        requirements: [
          requirement({
            id: "procedure:alpha",
            reason: "Beta requires alpha.",
          }),
        ],
      }),
    ]);

    expect(
      compositionCatalogErrors(cyclic, [
        installed("alpha-guide"),
        installed("beta-guide"),
      ]),
    ).toContain(
      "required semantic dependency cycle alpha-guide → beta-guide → alpha-guide",
    );
  });
});

describe("installed-snapshot selection", () => {
  it("selects an exact installed version and records provenance", () => {
    const result = resolveExactSkill(
      catalog([entry("alpha-guide")]),
      [installed("alpha-guide", "7.0.0")],
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

  it("does not substitute a different installed implementation version", () => {
    const result = resolveExactSkill(
      catalog([entry("alpha-guide")]),
      [installed("alpha-guide", "7.0.0")],
      "alpha-guide",
      "6.0.0",
      {
        kind: "exact-dependency",
        reason: "A manifest dependency pinned the implementation.",
        requestedBy: "consumer-guide",
      },
    );

    expect(result).toMatchObject({
      status: "incompatible",
      blocking: true,
      candidates: ["alpha-guide@7.0.0"],
    });
  });

  it("selects the sole compatible installed semantic provision", () => {
    const current = catalog([
      entry("alpha-guide"),
      entry("beta-guide", {
        provisions: [{id: "assurance:evidence", version: "1.2.0"}],
      }),
      entry("gamma-guide", {
        provisions: [{id: "assurance:evidence", version: "2.0.0"}],
      }),
    ]);
    const result = resolveSemanticRequirement(
      current,
      [
        installed("alpha-guide"),
        installed("beta-guide", "7.0.0"),
        installed("gamma-guide", "7.0.0"),
      ],
      requirement(),
      {
        kind: "semantic-requirement",
        reason: "The selected procedure requires evidence.",
        requestedBy: "alpha-guide",
      },
    );

    expect(result).toMatchObject({
      status: "selected",
      selection: {
        guide: "beta-guide",
        implementationVersion: "7.0.0",
        provision: {
          id: "assurance:evidence",
          version: "1.2.0",
        },
        requestedRange: "^1.0.0",
        provenance: {
          kind: "semantic-requirement",
          requestedBy: "alpha-guide",
        },
      },
    });
  });

  it("reports incompatible and ambiguous providers without choosing", () => {
    const incompatible = catalog([
      entry("beta-guide", {
        provisions: [{id: "assurance:evidence", version: "2.0.0"}],
      }),
    ]);
    expect(
      resolveSemanticRequirement(
        incompatible,
        [installed("beta-guide")],
        requirement(),
        {
          kind: "semantic-requirement",
          reason: "Evidence is required.",
        },
      ),
    ).toMatchObject({
      status: "incompatible",
      blocking: true,
      candidates: ["beta-guide@2.0.0"],
    });

    const ambiguous = catalog([
      entry("beta-guide", {
        provisions: [{id: "assurance:evidence", version: "1.0.0"}],
      }),
      entry("gamma-guide", {
        provisions: [{id: "assurance:evidence", version: "1.1.0"}],
      }),
    ]);
    expect(
      resolveSemanticRequirement(
        ambiguous,
        [installed("beta-guide"), installed("gamma-guide")],
        requirement(),
        {
          kind: "policy",
          reason: "Policy requires evidence assurance.",
        },
      ),
    ).toMatchObject({
      status: "ambiguous",
      blocking: true,
      candidates: ["beta-guide@1.0.0", "gamma-guide@1.1.0"],
    });
  });
});
