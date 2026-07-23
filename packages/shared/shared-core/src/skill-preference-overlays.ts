import {
  type CompositionCatalog,
  type InstalledSkill,
  type OverlayContext,
  type OverlayScope,
  type PreferenceOverlayRequest,
} from "./skill-composition-contract.js";
import {
  resolveExactSkill,
  selectionFailure,
  type FailedSelection,
  type SelectedSkill,
  type SelectionProvenance,
} from "./skill-selection.js";

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
  readonly scopes: readonly OverlayScope[];
  readonly status: "conflict";
  readonly target: string;
}

export interface SkippedPreferenceOverlay {
  readonly guide: string;
  readonly reason: string;
  readonly scope: OverlayScope;
  readonly target: string;
}

export interface PreferenceOverlayResolution {
  readonly applied: readonly SelectedPreferenceOverlay[];
  readonly conflicts: readonly PreferenceOverlayConflict[];
  readonly failures: readonly FailedSelection[];
  readonly shadowed: readonly SelectedPreferenceOverlay[];
  readonly skipped: readonly SkippedPreferenceOverlay[];
}

const DEFAULT_OVERLAY_CONTEXT: OverlayContext = {
  global: "*",
  languages: [],
  frameworks: [],
  paths: [],
  explicit: [],
};

const overlayApplies = (
  scope: OverlayScope,
  context: OverlayContext,
): boolean => {
  switch (scope.kind) {
    case "global": {
      return scope.value === "*" || scope.value === context.global;
    }
    case "organization": {
      return scope.value === context.organization;
    }
    case "repository": {
      return scope.value === context.repository;
    }
    case "language": {
      return context.languages.includes(scope.value);
    }
    case "framework": {
      return context.frameworks.includes(scope.value);
    }
    case "path": {
      return context.paths.some(
        (path) => path === scope.value || path.startsWith(`${scope.value}/`),
      );
    }
    case "explicit": {
      return scope.value === "*" || context.explicit.includes(scope.value);
    }
  }
};

const scopeSpecificity = (
  left: OverlayScope,
  right: OverlayScope,
): -1 | 0 | 1 | undefined => {
  if (left.kind === right.kind) {
    if (left.value === right.value) return 0;
    if (left.kind !== "path") return undefined;
    if (right.value.startsWith(`${left.value}/`)) return -1;
    if (left.value.startsWith(`${right.value}/`)) return 1;
    return undefined;
  }
  if (left.kind === "global" || right.kind === "explicit") return -1;
  if (right.kind === "global" || left.kind === "explicit") return 1;
  const orderedPairs = new Set([
    "organization\0repository",
    "organization\0path",
    "repository\0path",
    "language\0framework",
  ]);
  if (orderedPairs.has(`${left.kind}\0${right.kind}`)) return -1;
  if (orderedPairs.has(`${right.kind}\0${left.kind}`)) return 1;
  return undefined;
};

const compareOverlay = (
  left: SelectedPreferenceOverlay,
  right: SelectedPreferenceOverlay,
): number =>
  left.precedence - right.precedence ||
  left.scope.value.localeCompare(right.scope.value) ||
  left.guide.localeCompare(right.guide);

export const resolvePreferenceOverlays = (
  catalog: CompositionCatalog,
  installedSkills: readonly InstalledSkill[],
  overlays: readonly PreferenceOverlayRequest[],
  context: OverlayContext = DEFAULT_OVERLAY_CONTEXT,
): PreferenceOverlayResolution => {
  const failures: FailedSelection[] = [];
  const selected: SelectedPreferenceOverlay[] = [];
  const skipped: SkippedPreferenceOverlay[] = [];
  for (const overlay of overlays) {
    if (!overlayApplies(overlay.scope, context)) {
      skipped.push(overlay);
      continue;
    }
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
        selectionFailure(catalog, provenance, {
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
  const applied: SelectedPreferenceOverlay[] = [];
  const shadowed: SelectedPreferenceOverlay[] = [];
  const unique = selected.filter(
    (overlay, index, values) =>
      values.findIndex(
        ({guide, scope, target}) =>
          guide === overlay.guide &&
          target === overlay.target &&
          scope.kind === overlay.scope.kind &&
          scope.value === overlay.scope.value,
      ) === index,
  );
  for (const [target, group] of Map.groupBy(unique, ({target}) => target)) {
    const maximal = group.filter((candidate) =>
      group.every(
        (other) => scopeSpecificity(candidate.scope, other.scope) !== -1,
      ),
    );
    const rawConflictMembers: SelectedPreferenceOverlay[] = [];
    for (const [index, left] of maximal.entries()) {
      for (const right of maximal.slice(index + 1)) {
        const relation = scopeSpecificity(left.scope, right.scope);
        if (
          left.guide !== right.guide &&
          (relation === undefined || relation === 0)
        ) {
          rawConflictMembers.push(left, right);
        }
      }
    }
    const conflictMembers = rawConflictMembers.filter(
      (overlay, index, values) =>
        values.findIndex(
          ({guide, scope}) =>
            guide === overlay.guide &&
            scope.kind === overlay.scope.kind &&
            scope.value === overlay.scope.value,
        ) === index,
    );
    if (conflictMembers.length > 0) {
      const first = conflictMembers[0];
      if (first === undefined) continue;
      const scopes = conflictMembers.map(({scope}) => scope);
      conflicts.push({
        status: "conflict",
        candidates: [
          ...new Set(conflictMembers.map(({guide}) => guide)),
        ].toSorted(),
        message: `${target} has preference overlays with equal or incomparable applicable scopes`,
        scope: first.scope,
        scopes,
        target,
      });
      shadowed.push(...group.filter((overlay) => !maximal.includes(overlay)));
      continue;
    }

    const orderedMaximal = maximal.toSorted(compareOverlay);
    const effective = orderedMaximal.at(-1);
    if (effective === undefined) continue;
    applied.push(effective);
    shadowed.push(...group.filter((overlay) => overlay !== effective));
  }
  return {
    applied: applied.toSorted(compareOverlay),
    conflicts: conflicts.toSorted((left, right) =>
      `${left.target}\0${left.scope.kind}\0${left.scope.value}`.localeCompare(
        `${right.target}\0${right.scope.kind}\0${right.scope.value}`,
      ),
    ),
    failures,
    shadowed: shadowed.toSorted(compareOverlay),
    skipped: skipped.toSorted((left, right) =>
      `${left.target}\0${left.scope.kind}\0${left.scope.value}\0${left.guide}`.localeCompare(
        `${right.target}\0${right.scope.kind}\0${right.scope.value}\0${right.guide}`,
      ),
    ),
  };
};
