import {
  type CompositionCatalog,
  type InstalledSkill,
} from "./skill-composition-contract.js";
import {
  resolveSemanticRequirement,
  type FailedSelection,
} from "./skill-selection.js";

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
