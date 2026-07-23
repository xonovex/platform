import {createHash} from "node:crypto";
import {major, satisfies, valid, validRange} from "semver";
import {z} from "zod";

const GUIDE_NAME_RE = /^[a-z0-9]+(?:-[a-z0-9]+)*-guide$/;
const PROVISION_ID_RE = /^[a-z][a-z0-9-]*:[a-z0-9]+(?:[.-][a-z0-9]+)*$/;
const SUPPORTED_CONTRACT_MAJOR = 2;

export const DEFAULT_OVERLAY_PRECEDENCE = [
  "global",
  "organization",
  "repository",
  "language",
  "framework",
  "path",
  "explicit",
] as const;

const LifecycleSchema = z.enum(["durable", "procedural"]);
const FunctionalRoleSchema = z.enum([
  "domain",
  "context",
  "preference",
  "procedure",
  "capability-use",
  "assurance",
  "recovery",
  "communication",
]);
const ProvisionSchema = z
  .object({
    id: z.string().regex(PROVISION_ID_RE),
    version: z.string().refine((value) => valid(value) !== null, {
      message: "must be a valid semantic version",
    }),
  })
  .strict();
const RequirementSchema = z
  .object({
    id: z.string().regex(PROVISION_ID_RE),
    range: z.string().refine((value) => validRange(value) !== null, {
      message: "must be a valid semantic-version range",
    }),
    strength: z.enum(["required", "preferred"]),
    reason: z.string().trim().min(1),
  })
  .strict();
const OverlayScopeKindSchema = z.enum(DEFAULT_OVERLAY_PRECEDENCE);
const OverlayScopeSchema = z
  .object({
    kind: OverlayScopeKindSchema,
    value: z.string().trim().min(1),
  })
  .strict();
const CatalogEntrySchema = z
  .object({
    name: z.string().regex(GUIDE_NAME_RE),
    classification: z
      .object({
        lifecycle: LifecycleSchema,
        functionalRole: FunctionalRoleSchema,
      })
      .strict(),
    provisions: z.array(ProvisionSchema).default([]),
    requirements: z.array(RequirementSchema).default([]),
  })
  .strict();
const OverlayPrecedenceSchema = z
  .array(OverlayScopeKindSchema)
  .length(DEFAULT_OVERLAY_PRECEDENCE.length)
  .refine(
    (values) =>
      new Set(values).size === DEFAULT_OVERLAY_PRECEDENCE.length &&
      DEFAULT_OVERLAY_PRECEDENCE.every((value) => values.includes(value)),
    {message: "must contain every overlay scope exactly once"},
  );
const CatalogFileSchema = z
  .object({
    contractVersion: z.string().refine((value) => valid(value) !== null, {
      message: "must be a valid semantic version",
    }),
    overlayPrecedence: OverlayPrecedenceSchema.default([
      ...DEFAULT_OVERLAY_PRECEDENCE,
    ]),
    skills: z.array(CatalogEntrySchema),
  })
  .strict();
const InstalledSkillSchema = z
  .object({
    guide: z.string().regex(GUIDE_NAME_RE),
    implementationVersion: z.string().refine((value) => valid(value) !== null, {
      message: "must be a valid semantic version",
    }),
    packagePath: z.string().trim().min(1).optional(),
    plugin: z.string().trim().min(1),
    sourcesPath: z.string().trim().min(1).optional(),
  })
  .strict();
const InstalledSkillInventorySchema = z
  .array(InstalledSkillSchema)
  .superRefine((skills, context) => {
    const firstIndexByGuide = new Map<string, number>();
    for (const [index, skill] of skills.entries()) {
      const firstIndex = firstIndexByGuide.get(skill.guide);
      if (firstIndex === undefined) {
        firstIndexByGuide.set(skill.guide, index);
        continue;
      }
      context.addIssue({
        code: "custom",
        message: `duplicates installed guide at index ${String(firstIndex)}`,
        path: [index, "guide"],
      });
    }
  });
const ExactSkillRequestSchema = z
  .object({
    guide: z.string().regex(GUIDE_NAME_RE),
    implementationVersion: z
      .string()
      .refine((value) => valid(value) !== null, {
        message: "must be a valid semantic version",
      })
      .optional(),
    reason: z.string().trim().min(1),
    required: z.boolean().default(true),
  })
  .strict();
const PreferenceOverlayRequestSchema = z
  .object({
    guide: z.string().regex(GUIDE_NAME_RE),
    reason: z.string().trim().min(1),
    scope: OverlayScopeSchema,
    target: z.string().trim().min(1),
  })
  .strict();
const CompositionRequestSchema = z
  .object({
    exactSkills: z.array(ExactSkillRequestSchema).default([]),
    preferenceOverlays: z.array(PreferenceOverlayRequestSchema).default([]),
    requirements: z.array(RequirementSchema).default([]),
  })
  .strict();

export type Lifecycle = z.infer<typeof LifecycleSchema>;
export type FunctionalRole = z.infer<typeof FunctionalRoleSchema>;
export type SemanticProvision = z.infer<typeof ProvisionSchema>;
export type SemanticRequirement = z.infer<typeof RequirementSchema>;
export type OverlayScopeKind = z.infer<typeof OverlayScopeKindSchema>;
export type OverlayScope = z.infer<typeof OverlayScopeSchema>;
export type CompositionCatalogEntry = z.infer<typeof CatalogEntrySchema>;
export type InstalledSkill = z.infer<typeof InstalledSkillSchema>;
export type ExactSkillRequest = z.infer<typeof ExactSkillRequestSchema>;
export type PreferenceOverlayRequest = z.infer<
  typeof PreferenceOverlayRequestSchema
>;
export type CompositionRequest = z.infer<typeof CompositionRequestSchema>;

export interface CompositionCatalog {
  readonly contractVersion: string;
  readonly digest: string;
  readonly overlayPrecedence: readonly OverlayScopeKind[];
  readonly skills: readonly CompositionCatalogEntry[];
}

export interface SelectionProvenance {
  readonly kind:
    "explicit" | "exact-dependency" | "semantic-requirement" | "policy";
  readonly reason: string;
  readonly requestedBy?: string;
}

export interface SelectedSkill {
  readonly catalogContractVersion: string;
  readonly catalogDigest: string;
  readonly guide: string;
  readonly implementationVersion: string;
  readonly packagePath?: string;
  readonly plugin: string;
  readonly provenance: SelectionProvenance;
  readonly provision?: SemanticProvision;
  readonly requestedRange?: string;
  readonly sourcesPath?: string;
}

export interface FailedSelection {
  readonly blocking: boolean;
  readonly candidates: readonly string[];
  readonly catalogContractVersion: string;
  readonly catalogDigest: string;
  readonly message: string;
  readonly provenance: SelectionProvenance;
  readonly requestedGuide?: string;
  readonly requestedRange?: string;
  readonly requestedVersion?: string;
  readonly requirement?: SemanticRequirement;
  readonly status: "ambiguous" | "conflict" | "incompatible" | "unavailable";
}

export type SelectionResult =
  | {
      readonly status: "selected";
      readonly selection: SelectedSkill;
    }
  | FailedSelection;

export type SkillSelectionRecord =
  | (SelectedSkill & {
      readonly reason: string;
      readonly required: boolean;
      readonly status: "selected";
    })
  | (FailedSelection & {
      readonly reason: string;
      readonly required: boolean;
    });

export interface SelectedPreferenceOverlay {
  readonly guide: string;
  readonly precedence: number;
  readonly reason: string;
  readonly scope: OverlayScope;
  readonly selection: SelectedSkill;
  readonly target: string;
}

export interface PreferenceOverlayConflict {
  readonly candidates: readonly string[];
  readonly message: string;
  readonly scope: OverlayScope;
  readonly status: "conflict";
  readonly target: string;
}

export interface PreferenceOverlayResolution {
  readonly applied: readonly SelectedPreferenceOverlay[];
  readonly conflicts: readonly PreferenceOverlayConflict[];
  readonly failures: readonly FailedSelection[];
}

export interface CompositionResolution {
  readonly catalogContractVersion: string;
  readonly catalogDigest: string;
  readonly overlays: PreferenceOverlayResolution;
  readonly skills: readonly SkillSelectionRecord[];
  readonly status: "blocked" | "degraded" | "ready";
}

type ParseResult<T> =
  | {readonly success: true; readonly data: T}
  | {readonly success: false; readonly errors: readonly string[]};

const issueDetail = (issue: z.core.$ZodIssue): string => {
  const path =
    issue.path.length > 0 ? issue.path.map(String).join(".") : "root";
  return `${path}: ${issue.message}`;
};

const parseWithSchema = <T>(
  schema: z.ZodType<T>,
  input: unknown,
): ParseResult<T> => {
  const parsed = schema.safeParse(input);
  return parsed.success
    ? {success: true, data: parsed.data}
    : {success: false, errors: parsed.error.issues.map(issueDetail)};
};

export const parseCompositionCatalog = (
  input: unknown,
  sourceText = JSON.stringify(input),
): ParseResult<CompositionCatalog> => {
  const parsed = CatalogFileSchema.safeParse(input);
  if (!parsed.success) {
    return {
      success: false,
      errors: parsed.error.issues.map(issueDetail),
    };
  }
  if (major(parsed.data.contractVersion) !== SUPPORTED_CONTRACT_MAJOR) {
    return {
      success: false,
      errors: [
        `contractVersion: unsupported major ${String(major(parsed.data.contractVersion))}; expected ${String(SUPPORTED_CONTRACT_MAJOR)}`,
      ],
    };
  }
  return {
    success: true,
    data: {
      ...parsed.data,
      digest: createHash("sha256").update(sourceText).digest("hex"),
    },
  };
};

export const parseInstalledSkillInventory = (
  input: unknown,
): ParseResult<readonly InstalledSkill[]> =>
  parseWithSchema(InstalledSkillInventorySchema, input);

export const parseCompositionRequest = (
  input: unknown,
): ParseResult<CompositionRequest> =>
  parseWithSchema(CompositionRequestSchema, input);

const selectedSkill = (
  catalog: CompositionCatalog,
  installed: InstalledSkill,
  provenance: SelectionProvenance,
  provision?: SemanticProvision,
  requestedRange?: string,
): SelectionResult => ({
  status: "selected",
  selection: {
    catalogContractVersion: catalog.contractVersion,
    catalogDigest: catalog.digest,
    guide: installed.guide,
    implementationVersion: installed.implementationVersion,
    ...(installed.packagePath === undefined
      ? {}
      : {packagePath: installed.packagePath}),
    plugin: installed.plugin,
    provenance,
    ...(provision === undefined ? {} : {provision}),
    ...(requestedRange === undefined ? {} : {requestedRange}),
    ...(installed.sourcesPath === undefined
      ? {}
      : {sourcesPath: installed.sourcesPath}),
  },
});

const failedSelection = (
  catalog: CompositionCatalog,
  provenance: SelectionProvenance,
  failure: Omit<
    FailedSelection,
    "catalogContractVersion" | "catalogDigest" | "provenance"
  >,
): FailedSelection => ({
  ...failure,
  catalogContractVersion: catalog.contractVersion,
  catalogDigest: catalog.digest,
  provenance,
});

export const resolveExactSkill = (
  catalog: CompositionCatalog,
  installedSkills: readonly InstalledSkill[],
  guide: string,
  implementationVersion: string | undefined,
  provenance: SelectionProvenance,
  blocking = true,
): SelectionResult => {
  if (catalog.skills.every((entry) => entry.name !== guide)) {
    return failedSelection(catalog, provenance, {
      status: "unavailable",
      blocking,
      candidates: [],
      message: `exact skill ${guide} is not declared in catalog ${catalog.contractVersion}`,
      requestedGuide: guide,
      ...(implementationVersion === undefined
        ? {}
        : {requestedVersion: implementationVersion}),
    });
  }
  const installed = installedSkills.filter((skill) => skill.guide === guide);
  if (installed.length === 0) {
    return failedSelection(catalog, provenance, {
      status: "unavailable",
      blocking,
      candidates: [],
      message: `exact skill ${guide} is not installed`,
      requestedGuide: guide,
      ...(implementationVersion === undefined
        ? {}
        : {requestedVersion: implementationVersion}),
    });
  }
  if (installed.length > 1) {
    return failedSelection(catalog, provenance, {
      status: "ambiguous",
      blocking,
      candidates: installed
        .map(
          (skill) =>
            `${skill.plugin}:${skill.guide}@${skill.implementationVersion}`,
        )
        .toSorted(),
      message: `exact skill ${guide} has multiple installed identities`,
      requestedGuide: guide,
      ...(implementationVersion === undefined
        ? {}
        : {requestedVersion: implementationVersion}),
    });
  }
  const selected = installed[0];
  if (selected === undefined) {
    throw new Error("installed exact selection disappeared");
  }
  if (
    implementationVersion !== undefined &&
    selected.implementationVersion !== implementationVersion
  ) {
    return failedSelection(catalog, provenance, {
      status: "incompatible",
      blocking,
      candidates: [`${guide}@${selected.implementationVersion}`],
      message: `exact skill ${guide}@${implementationVersion} does not match installed ${selected.implementationVersion}`,
      requestedGuide: guide,
      requestedVersion: implementationVersion,
    });
  }
  return selectedSkill(catalog, selected, provenance);
};

export const resolveSemanticRequirement = (
  catalog: CompositionCatalog,
  installedSkills: readonly InstalledSkill[],
  requirement: SemanticRequirement,
  provenance: SelectionProvenance,
): SelectionResult => {
  const installedByGuide = new Map(
    installedSkills.map((skill) => [skill.guide, skill]),
  );
  const declared = catalog.skills.flatMap((entry) =>
    entry.provisions
      .filter((provision) => provision.id === requirement.id)
      .map((provision) => ({entry, provision})),
  );
  const declaredGuides = new Set(declared.map(({entry}) => entry.name));
  const duplicateInstalledGuides = installedSkills
    .map(({guide}) => guide)
    .filter(
      (guide, index, guides) =>
        declaredGuides.has(guide) &&
        guides.indexOf(guide) !== index &&
        guides.lastIndexOf(guide) === index,
    );
  if (duplicateInstalledGuides.length > 0) {
    return failedSelection(catalog, provenance, {
      status: "ambiguous",
      blocking: requirement.strength === "required",
      candidates: installedSkills
        .filter(({guide}) => duplicateInstalledGuides.includes(guide))
        .map(
          ({guide, implementationVersion, plugin}) =>
            `${plugin}:${guide}@${implementationVersion}`,
        )
        .toSorted(),
      message: `${requirement.id}@${requirement.range} has providers with duplicate installed identities`,
      requestedRange: requirement.range,
      requirement,
    });
  }
  const installed = declared.filter(({entry}) =>
    installedByGuide.has(entry.name),
  );
  const compatible = installed.filter(({provision}) =>
    satisfies(provision.version, requirement.range),
  );
  const blocking = requirement.strength === "required";

  if (compatible.length === 0) {
    const candidates = installed.map(
      ({entry, provision}) => `${entry.name}@${provision.version}`,
    );
    return failedSelection(catalog, provenance, {
      status: candidates.length > 0 ? "incompatible" : "unavailable",
      blocking,
      candidates,
      message:
        candidates.length > 0
          ? `${requirement.id} has no installed provision compatible with ${requirement.range}`
          : `${requirement.id} has no installed provider`,
      requestedRange: requirement.range,
      requirement,
    });
  }
  if (compatible.length > 1) {
    return failedSelection(catalog, provenance, {
      status: "ambiguous",
      blocking,
      candidates: compatible
        .map(({entry, provision}) => `${entry.name}@${provision.version}`)
        .toSorted(),
      message: `${requirement.id}@${requirement.range} has multiple compatible installed providers`,
      requestedRange: requirement.range,
      requirement,
    });
  }
  const candidate = compatible[0];
  if (candidate === undefined) {
    throw new Error("compatible semantic selection disappeared");
  }
  const installedCandidate = installedByGuide.get(candidate.entry.name);
  if (installedCandidate === undefined) {
    throw new Error("installed semantic selection disappeared");
  }
  return selectedSkill(
    catalog,
    installedCandidate,
    provenance,
    candidate.provision,
    requirement.range,
  );
};

const dependencyCycles = (
  edges: ReadonlyMap<string, ReadonlySet<string>>,
): readonly string[] => {
  const states = new Map<string, "visiting" | "visited">();
  const cycles = new Set<string>();
  const visit = (name: string, path: readonly string[]): void => {
    if (states.get(name) === "visited") return;
    if (states.get(name) === "visiting") {
      const start = path.indexOf(name);
      cycles.add([...path.slice(start), name].join(" → "));
      return;
    }
    states.set(name, "visiting");
    for (const dependency of edges.get(name) ?? []) {
      visit(dependency, [...path, name]);
    }
    states.set(name, "visited");
  };
  for (const name of edges.keys()) visit(name, []);
  return [...cycles].toSorted();
};

export const compositionCatalogIssues = (
  catalog: CompositionCatalog,
  installedSkills: readonly InstalledSkill[],
): {
  readonly errors: readonly string[];
  readonly preferredFailures: readonly FailedSelection[];
} => {
  const errors: string[] = [];
  const preferredFailures: FailedSelection[] = [];
  const names = catalog.skills.map(({name}) => name);
  const duplicateNames = names.filter(
    (name, index) => names.indexOf(name) !== index,
  );
  for (const name of new Set(duplicateNames)) {
    errors.push(`duplicate skill entry ${name}`);
  }
  if (
    names.some((name, index) => index > 0 && name < (names[index - 1] ?? ""))
  ) {
    errors.push("skill entries must be sorted alphabetically by name");
  }

  const installedNames = new Set(installedSkills.map(({guide}) => guide));
  const catalogNames = new Set(names);
  const installedGuideNames = installedSkills.map(({guide}) => guide);
  for (const duplicate of new Set(
    installedGuideNames.filter(
      (guide, index) => installedGuideNames.indexOf(guide) !== index,
    ),
  )) {
    errors.push(`installed snapshot contains duplicate guide ${duplicate}`);
  }
  for (const missing of [...installedNames]
    .filter((name) => !catalogNames.has(name))
    .toSorted()) {
    errors.push(`missing installed skill ${missing}`);
  }
  for (const unexpected of [...catalogNames]
    .filter((name) => !installedNames.has(name))
    .toSorted()) {
    errors.push(`catalog skill ${unexpected} is not installed`);
  }

  const edges = new Map<string, Set<string>>();
  for (const entry of catalog.skills) {
    const provisionIds = entry.provisions.map(({id}) => id);
    for (const duplicate of new Set(
      provisionIds.filter((id, index) => provisionIds.indexOf(id) !== index),
    )) {
      errors.push(
        `${entry.name} declares provision ${duplicate} more than once`,
      );
    }
    const requirementKeys = entry.requirements.map(({id}) => id);
    for (const duplicate of new Set(
      requirementKeys.filter(
        (key, index) => requirementKeys.indexOf(key) !== index,
      ),
    )) {
      errors.push(
        `${entry.name} declares requirement ${duplicate} more than once`,
      );
    }

    for (const requirement of entry.requirements) {
      const resolution = resolveSemanticRequirement(
        catalog,
        installedSkills,
        requirement,
        {
          kind: "semantic-requirement",
          reason: requirement.reason,
          requestedBy: entry.name,
        },
      );
      if (resolution.status !== "selected") {
        if (requirement.strength === "preferred") {
          preferredFailures.push(resolution);
        } else {
          errors.push(
            `${entry.name} required requirement is not deterministic: ${resolution.message}`,
          );
        }
        continue;
      }
      if (requirement.strength !== "required") continue;
      const dependencies = edges.get(entry.name) ?? new Set<string>();
      dependencies.add(resolution.selection.guide);
      edges.set(entry.name, dependencies);
    }
  }
  for (const cycle of dependencyCycles(edges)) {
    errors.push(`required semantic dependency cycle ${cycle}`);
  }
  return {errors, preferredFailures};
};

export const compositionCatalogSnapshotErrors = (
  canonical: string,
  snapshot: string | undefined,
): readonly string[] => {
  if (snapshot === undefined) {
    return [
      "packaged workflow snapshot workflow-guide/assets/composition-catalog.json is missing",
    ];
  }
  if (snapshot !== canonical) {
    return [
      "packaged workflow snapshot differs from packages/skill/composition-catalog.json",
    ];
  }
  return [];
};

export const resolvePreferenceOverlays = (
  catalog: CompositionCatalog,
  installedSkills: readonly InstalledSkill[],
  overlays: readonly PreferenceOverlayRequest[],
): PreferenceOverlayResolution => {
  const failures: FailedSelection[] = [];
  const selected: SelectedPreferenceOverlay[] = [];
  for (const overlay of overlays) {
    const provenance: SelectionProvenance = {
      kind: "policy",
      reason: overlay.reason,
      requestedBy: `overlay:${overlay.target}`,
    };
    const exact = resolveExactSkill(
      catalog,
      installedSkills,
      overlay.guide,
      undefined,
      provenance,
      true,
    );
    if (exact.status !== "selected") {
      failures.push(exact);
      continue;
    }
    const catalogEntry = catalog.skills.find(
      (entry) => entry.name === overlay.guide,
    );
    if (catalogEntry?.classification.functionalRole !== "preference") {
      failures.push(
        failedSelection(catalog, provenance, {
          status: "conflict",
          blocking: true,
          candidates: [overlay.guide],
          message: `${overlay.guide} is ${catalogEntry?.classification.functionalRole ?? "unclassified"} and cannot be applied as a preference overlay`,
          requestedGuide: overlay.guide,
        }),
      );
      continue;
    }
    selected.push({
      guide: overlay.guide,
      precedence: catalog.overlayPrecedence.indexOf(overlay.scope.kind),
      reason: overlay.reason,
      scope: overlay.scope,
      selection: exact.selection,
      target: overlay.target,
    });
  }

  const conflicts: PreferenceOverlayConflict[] = [];
  const conflictedKeys = new Set<string>();
  const groups = Map.groupBy(
    selected,
    ({scope, target}) => `${target}\0${scope.kind}\0${scope.value}`,
  );
  for (const [key, group] of groups) {
    const candidates = [...new Set(group.map(({guide}) => guide))].toSorted();
    if (candidates.length < 2) continue;
    conflictedKeys.add(key);
    const first = group[0];
    if (first === undefined) continue;
    conflicts.push({
      status: "conflict",
      candidates,
      message: `${first.target} has multiple preference overlays at equal ${first.scope.kind} scope ${first.scope.value}`,
      scope: first.scope,
      target: first.target,
    });
  }

  const applied = selected
    .filter(
      ({scope, target}) =>
        !conflictedKeys.has(`${target}\0${scope.kind}\0${scope.value}`),
    )
    .filter(
      (overlay, index, values) =>
        values.findIndex(
          ({guide, scope, target}) =>
            guide === overlay.guide &&
            target === overlay.target &&
            scope.kind === overlay.scope.kind &&
            scope.value === overlay.scope.value,
        ) === index,
    )
    .toSorted(
      (left, right) =>
        left.precedence - right.precedence ||
        left.scope.value.localeCompare(right.scope.value) ||
        left.guide.localeCompare(right.guide),
    );
  return {
    applied,
    conflicts: conflicts.toSorted((left, right) =>
      `${left.target}\0${left.scope.kind}\0${left.scope.value}`.localeCompare(
        `${right.target}\0${right.scope.kind}\0${right.scope.value}`,
      ),
    ),
    failures,
  };
};

const selectionRecord = (
  result: SelectionResult,
  reason: string,
  required: boolean,
): SkillSelectionRecord =>
  result.status === "selected"
    ? {
        ...result.selection,
        reason,
        required,
        status: "selected",
      }
    : {...result, reason, required};

const compositionStatus = (
  blocked: boolean,
  degraded: boolean,
): CompositionResolution["status"] => {
  if (blocked) return "blocked";
  if (degraded) return "degraded";
  return "ready";
};

export const resolveComposition = (
  catalog: CompositionCatalog,
  installedSkills: readonly InstalledSkill[],
  request: CompositionRequest,
): CompositionResolution => {
  const exactResults: SkillSelectionRecord[] = [];
  const semanticResults: SkillSelectionRecord[] = [];
  const expandedGuides = new Set<string>();
  const resolveRequirement = (
    requirement: SemanticRequirement,
    requestedBy: string,
  ): SelectionResult =>
    resolveSemanticRequirement(catalog, installedSkills, requirement, {
      kind: "semantic-requirement",
      reason: requirement.reason,
      requestedBy,
    });
  const expandInvariantRequirements = (guide: string): void => {
    if (expandedGuides.has(guide)) return;
    expandedGuides.add(guide);
    const requirements =
      catalog.skills.find(({name}) => name === guide)?.requirements ?? [];
    for (const requirement of requirements) {
      const result = resolveRequirement(requirement, guide);
      semanticResults.push(
        selectionRecord(
          result,
          requirement.reason,
          requirement.strength === "required",
        ),
      );
      if (result.status === "selected") {
        expandInvariantRequirements(result.selection.guide);
      }
    }
  };

  for (const exact of request.exactSkills) {
    const result = resolveExactSkill(
      catalog,
      installedSkills,
      exact.guide,
      exact.implementationVersion,
      {
        kind: "explicit",
        reason: exact.reason,
        requestedBy: "composition-request",
      },
      exact.required,
    );
    exactResults.push(selectionRecord(result, exact.reason, exact.required));
    if (result.status === "selected") {
      expandInvariantRequirements(result.selection.guide);
    }
  }
  for (const requirement of request.requirements) {
    const result = resolveRequirement(requirement, "composition-request");
    semanticResults.push(
      selectionRecord(
        result,
        requirement.reason,
        requirement.strength === "required",
      ),
    );
    if (result.status === "selected") {
      expandInvariantRequirements(result.selection.guide);
    }
  }
  const overlays = resolvePreferenceOverlays(
    catalog,
    installedSkills,
    request.preferenceOverlays,
  );
  for (const overlay of overlays.applied) {
    expandInvariantRequirements(overlay.guide);
  }
  const skills = [...exactResults, ...semanticResults];
  const blocked =
    skills.some(
      (selection) => selection.status !== "selected" && selection.blocking,
    ) ||
    overlays.conflicts.length > 0 ||
    overlays.failures.some(({blocking}) => blocking);
  const degraded =
    skills.some((selection) => selection.status !== "selected") ||
    overlays.failures.length > 0;
  return {
    catalogContractVersion: catalog.contractVersion,
    catalogDigest: catalog.digest,
    overlays,
    skills,
    status: compositionStatus(blocked, degraded),
  };
};
