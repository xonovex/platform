import {satisfies} from "semver";
import {
  type CompositionCatalog,
  type InstalledSkill,
  type SemanticProvision,
  type SemanticRequirement,
} from "./skill-composition-contract.js";

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

export const selectionFailure = (
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
    return selectionFailure(catalog, provenance, {
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
    return selectionFailure(catalog, provenance, {
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
    return selectionFailure(catalog, provenance, {
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
    return selectionFailure(catalog, provenance, {
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
  providerGuide?: string,
  selectedGuides: ReadonlySet<string> = new Set(),
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
    return selectionFailure(catalog, provenance, {
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
  const requestedProvider =
    providerGuide === undefined
      ? undefined
      : declared.find(({entry}) => entry.name === providerGuide);
  if (providerGuide !== undefined && requestedProvider === undefined) {
    return selectionFailure(catalog, provenance, {
      status: "incompatible",
      blocking,
      candidates: declared
        .map(({entry, provision}) => `${entry.name}@${provision.version}`)
        .toSorted(),
      message: `${providerGuide} does not provide ${requirement.id}`,
      requestedGuide: providerGuide,
      requestedRange: requirement.range,
      requirement,
    });
  }
  if (
    requestedProvider !== undefined &&
    !installedByGuide.has(requestedProvider.entry.name)
  ) {
    return selectionFailure(catalog, provenance, {
      status: "unavailable",
      blocking,
      candidates: [],
      message: `selected provider ${providerGuide ?? requestedProvider.entry.name} is not installed`,
      requestedGuide: requestedProvider.entry.name,
      requestedRange: requirement.range,
      requirement,
    });
  }
  if (
    requestedProvider !== undefined &&
    !satisfies(requestedProvider.provision.version, requirement.range)
  ) {
    return selectionFailure(catalog, provenance, {
      status: "incompatible",
      blocking,
      candidates: [
        `${requestedProvider.entry.name}@${requestedProvider.provision.version}`,
      ],
      message: `selected provider ${requestedProvider.entry.name} is incompatible with ${requirement.id}@${requirement.range}`,
      requestedGuide: requestedProvider.entry.name,
      requestedRange: requirement.range,
      requirement,
    });
  }

  if (compatible.length === 0) {
    const candidates = installed.map(
      ({entry, provision}) => `${entry.name}@${provision.version}`,
    );
    return selectionFailure(catalog, provenance, {
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
  const selectedCompatible = compatible.filter(({entry}) =>
    selectedGuides.has(entry.name),
  );
  let candidates = compatible;
  if (requestedProvider === undefined && selectedCompatible.length === 1) {
    candidates = selectedCompatible;
  } else if (requestedProvider !== undefined) {
    candidates = compatible.filter(({entry}) => entry.name === providerGuide);
  }
  if (candidates.length > 1) {
    return selectionFailure(catalog, provenance, {
      status: "ambiguous",
      blocking,
      candidates: candidates
        .map(({entry, provision}) => `${entry.name}@${provision.version}`)
        .toSorted(),
      message: `${requirement.id}@${requirement.range} has multiple compatible installed providers`,
      requestedRange: requirement.range,
      requirement,
    });
  }
  const candidate = candidates[0];
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
