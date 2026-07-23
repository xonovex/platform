import {
  type CompositionCatalog,
  type CompositionRequest,
  type InstalledSkill,
  type SemanticProvision,
  type SemanticRequirement,
} from "./skill-composition-contract.js";
import {
  resolvePreferenceOverlays,
  type PreferenceOverlayResolution,
} from "./skill-preference-overlays.js";
import {
  resolveExactSkill,
  resolveSemanticRequirement,
  selectionFailure,
  type FailedSelection,
  type SelectedSkill,
  type SelectionProvenance,
  type SelectionResult,
} from "./skill-selection.js";

export type SkillSelectionRecord =
  | (SelectedSkill & {
      readonly matchedProvisions: readonly {
        readonly provision: SemanticProvision;
        readonly requestedRange: string;
      }[];
      readonly reason: string;
      readonly reasons: readonly string[];
      readonly required: boolean;
      readonly status: "selected";
      readonly provenances: readonly SelectionProvenance[];
    })
  | (FailedSelection & {
      readonly reason: string;
      readonly required: boolean;
    });

export interface CompositionResolution {
  readonly catalogContractVersion: string;
  readonly catalogDigest: string;
  readonly loadOrder: readonly string[];
  readonly overlays: PreferenceOverlayResolution;
  readonly skills: readonly SkillSelectionRecord[];
  readonly status: "blocked" | "degraded" | "ready";
}

const provenancePriority = (provenance: SelectionProvenance): number => {
  switch (provenance.kind) {
    case "explicit": {
      return 0;
    }
    case "exact-dependency": {
      return 1;
    }
    case "policy": {
      return 2;
    }
    case "semantic-requirement": {
      return 3;
    }
  }
};

const compareRequirement = (
  left: SemanticRequirement,
  right: SemanticRequirement,
): number => {
  const strengthOrder = {required: 0, preferred: 1} as const;
  return (
    strengthOrder[left.strength] - strengthOrder[right.strength] ||
    left.id.localeCompare(right.id) ||
    left.range.localeCompare(right.range) ||
    left.reason.localeCompare(right.reason)
  );
};

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
  type SelectedRecord = Extract<
    SkillSelectionRecord,
    {readonly status: "selected"}
  >;
  const selectedRecords = new Map<string, SelectedRecord>();
  const failures: SkillSelectionRecord[] = [];
  const failureKeys = new Set<string>();
  const disambiguatingGuides = new Set<string>();
  const processingDependencies = new Set<string>();
  const loadOrder: string[] = [];
  const loadedGuides = new Set<string>();
  const expandedGuides = new Set<string>();
  const installedByGuide = new Map(
    installedSkills.map((skill) => [skill.guide, skill]),
  );
  const installedByPlugin = Map.groupBy(installedSkills, ({plugin}) => plugin);

  const addFailure = (
    failure: FailedSelection,
    reason: string,
    required: boolean,
  ): void => {
    const key = [
      failure.status,
      failure.requestedGuide ?? "",
      failure.requestedVersion ?? "",
      failure.requirement?.id ?? "",
      failure.requestedRange ?? "",
      failure.provenance.requestedBy ?? "",
      failure.message,
    ].join("\0");
    if (failureKeys.has(key)) return;
    failureKeys.add(key);
    failures.push({...failure, reason, required});
  };

  const mergeSelection = (
    selection: SelectedSkill,
    reason: string,
    required: boolean,
  ): void => {
    const existing = selectedRecords.get(selection.guide);
    const matched =
      selection.provision === undefined ||
      selection.requestedRange === undefined
        ? []
        : [
            {
              provision: selection.provision,
              requestedRange: selection.requestedRange,
            },
          ];
    if (existing === undefined) {
      selectedRecords.set(selection.guide, {
        ...selection,
        matchedProvisions: matched,
        provenance: selection.provenance,
        provenances: [selection.provenance],
        reason,
        reasons: [reason],
        required,
        status: "selected",
      });
      return;
    }
    const provenances = [...existing.provenances, selection.provenance].filter(
      (provenance, index, values) =>
        values.findIndex(
          (candidate) =>
            candidate.kind === provenance.kind &&
            candidate.reason === provenance.reason &&
            candidate.requestedBy === provenance.requestedBy,
        ) === index,
    );
    const matchedProvisions = [
      ...existing.matchedProvisions,
      ...matched,
    ].filter(
      (match, index, values) =>
        values.findIndex(
          (candidate) =>
            candidate.provision.id === match.provision.id &&
            candidate.provision.version === match.provision.version &&
            candidate.requestedRange === match.requestedRange,
        ) === index,
    );
    const primary = provenances.toSorted(
      (left, right) =>
        provenancePriority(left) - provenancePriority(right) ||
        left.reason.localeCompare(right.reason),
    )[0];
    if (primary === undefined) {
      throw new Error("selected skill lost its provenance");
    }
    selectedRecords.set(selection.guide, {
      ...existing,
      ...(selection.provision === undefined
        ? {}
        : {provision: selection.provision}),
      ...(selection.requestedRange === undefined
        ? {}
        : {requestedRange: selection.requestedRange}),
      matchedProvisions,
      provenance: primary,
      provenances,
      reason: primary.reason,
      reasons: [...new Set([...existing.reasons, reason])],
      required: existing.required || required,
    });
  };

  const dependencyFailure = (
    owner: InstalledSkill,
    plugin: string,
    implementationVersion: string | undefined,
    status: FailedSelection["status"],
    candidates: readonly string[],
    message: string,
  ): FailedSelection =>
    selectionFailure(
      catalog,
      {
        kind: "exact-dependency",
        reason: `${owner.guide} requires ${plugin}.`,
        requestedBy: owner.guide,
      },
      {
        status,
        blocking: true,
        candidates,
        message,
        ...(implementationVersion === undefined
          ? {}
          : {requestedVersion: implementationVersion}),
      },
    );

  const addSelectedGraph = (
    selection: SelectedSkill,
    reason: string,
    required: boolean,
    disambiguates: boolean,
  ): void => {
    mergeSelection(selection, reason, required);
    if (disambiguates) disambiguatingGuides.add(selection.guide);
    if (processingDependencies.has(selection.guide)) {
      addFailure(
        selectionFailure(catalog, selection.provenance, {
          status: "conflict",
          blocking: true,
          candidates: [...processingDependencies, selection.guide],
          message: `exact dependency cycle reaches ${selection.guide}`,
          requestedGuide: selection.guide,
        }),
        reason,
        true,
      );
      return;
    }
    if (loadedGuides.has(selection.guide)) return;

    processingDependencies.add(selection.guide);
    const installed = installedByGuide.get(selection.guide);
    for (const dependency of installed?.dependencies ?? []) {
      const candidates = installedByPlugin.get(dependency.plugin) ?? [];
      if (candidates.length === 0) {
        addFailure(
          dependencyFailure(
            installed ?? {
              ...selection,
              dependencies: [],
            },
            dependency.plugin,
            dependency.implementationVersion,
            "unavailable",
            [],
            `exact dependency ${dependency.plugin} required by ${selection.guide} is not installed`,
          ),
          `${selection.guide} requires ${dependency.plugin}.`,
          true,
        );
        continue;
      }
      if (candidates.length > 1) {
        addFailure(
          dependencyFailure(
            installed ?? {
              ...selection,
              dependencies: [],
            },
            dependency.plugin,
            dependency.implementationVersion,
            "ambiguous",
            candidates
              .map(
                ({guide, implementationVersion: version}) =>
                  `${guide}@${version}`,
              )
              .toSorted(),
            `exact dependency ${dependency.plugin} distributes multiple installed guides`,
          ),
          `${selection.guide} requires ${dependency.plugin}.`,
          true,
        );
        continue;
      }
      const candidate = candidates[0];
      if (candidate === undefined) continue;
      const dependencyResult = resolveExactSkill(
        catalog,
        installedSkills,
        candidate.guide,
        dependency.implementationVersion,
        {
          kind: "exact-dependency",
          reason: `${selection.guide} requires ${candidate.guide}.`,
          requestedBy: selection.guide,
        },
        true,
      );
      if (dependencyResult.status !== "selected") {
        addFailure(dependencyResult, dependencyResult.provenance.reason, true);
        continue;
      }
      addSelectedGraph(
        dependencyResult.selection,
        dependencyResult.selection.provenance.reason,
        true,
        true,
      );
    }
    processingDependencies.delete(selection.guide);
    loadedGuides.add(selection.guide);
    loadOrder.push(selection.guide);
  };

  const addResult = (
    result: SelectionResult,
    reason: string,
    required: boolean,
    disambiguates = false,
  ): void => {
    if (result.status === "selected") {
      addSelectedGraph(result.selection, reason, required, disambiguates);
      return;
    }
    addFailure(result, reason, required);
  };

  const resolveRequirement = (
    requirement: SemanticRequirement,
    requestedBy: string,
  ): SelectionResult =>
    resolveSemanticRequirement(
      catalog,
      installedSkills,
      requirement,
      {
        kind: "semantic-requirement",
        reason: requirement.reason,
        requestedBy,
      },
      request.requirementProviders[requirement.id],
      disambiguatingGuides,
    );
  const pendingRequired: {
    readonly requestedBy: string;
    readonly requirement: SemanticRequirement;
  }[] = [];
  const pendingPreferred: {
    readonly requestedBy: string;
    readonly requirement: SemanticRequirement;
  }[] = [];
  const enqueueRequirements = (
    requirements: readonly SemanticRequirement[],
    requestedBy: string,
  ): void => {
    for (const requirement of requirements.toSorted(compareRequirement)) {
      const pending = {requestedBy, requirement};
      if (requirement.strength === "required") {
        pendingRequired.push(pending);
      } else {
        pendingPreferred.push(pending);
      }
    }
  };
  const harvestInvariantRequirements = (): void => {
    let guide = loadOrder.find((name) => !expandedGuides.has(name));
    while (guide !== undefined) {
      expandedGuides.add(guide);
      const requirements =
        catalog.skills.find(({name}) => name === guide)?.requirements ?? [];
      enqueueRequirements(requirements, guide);
      guide = loadOrder.find((name) => !expandedGuides.has(name));
    }
  };
  const resolvePending = (pending: {
    readonly requestedBy: string;
    readonly requirement: SemanticRequirement;
  }): void => {
    const {requestedBy, requirement} = pending;
    const result = resolveRequirement(requirement, requestedBy);
    addResult(
      result,
      requirement.reason,
      requirement.strength === "required",
      true,
    );
  };
  const resolveRequiredClosure = (): void => {
    harvestInvariantRequirements();
    let pending = pendingRequired.shift();
    while (pending !== undefined) {
      resolvePending(pending);
      harvestInvariantRequirements();
      pending = pendingRequired.shift();
    }
  };

  enqueueRequirements(request.requirements, "composition-request");
  for (const exact of request.exactSkills.toSorted((left, right) =>
    left.guide.localeCompare(right.guide),
  )) {
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
    addResult(result, exact.reason, exact.required, true);
  }
  let pendingRootRequirement = pendingRequired.shift();
  while (pendingRootRequirement !== undefined) {
    resolvePending(pendingRootRequirement);
    pendingRootRequirement = pendingRequired.shift();
  }
  const overlays = resolvePreferenceOverlays(
    catalog,
    installedSkills,
    request.preferenceOverlays,
    request.overlayContext,
  );
  for (const overlay of overlays.applied) {
    addSelectedGraph(overlay.selection, overlay.reason, true, true);
  }
  resolveRequiredClosure();
  let pendingPreferredRequirement = pendingPreferred.shift();
  while (pendingPreferredRequirement !== undefined) {
    resolvePending(pendingPreferredRequirement);
    resolveRequiredClosure();
    pendingPreferredRequirement = pendingPreferred.shift();
  }
  const skills = [
    ...loadOrder.flatMap((guide) => {
      const record = selectedRecords.get(guide);
      return record === undefined ? [] : [record];
    }),
    ...failures,
  ];
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
    loadOrder,
    overlays,
    skills,
    status: compositionStatus(blocked, degraded),
  };
};
