import {checkIndependence} from "./independence-helpers.mjs";

const requiredModelBounds = [
  "fixedInputs",
  "outputSchema",
  "validator",
  "retryLimit",
  "tokenBudget",
  "timeoutSeconds",
];

const requiredAgentBounds = [
  "purpose",
  "resultContract",
  "toolScope",
  "filesystemScope",
  "networkScope",
  "maxDepth",
  "tokenBudget",
  "costBudget",
  "timeoutSeconds",
  "cancellation",
];

const hasValue = (value) => value !== undefined && value !== null;

const hasEveryValue = (value, fields) =>
  fields.every((field) => hasValue(value?.[field]));

const duplicateValues = (values) =>
  values.filter((value, index) => values.indexOf(value) !== index);

export const selectDevelopmentExecutor = ({
  workShape,
  requestedExecutor,
  bounds,
}) => {
  if (workShape === "mechanical") {
    return requestedExecutor && requestedExecutor !== "deterministic"
      ? {code: "least-adaptive-executor-bypassed"}
      : {executor: "deterministic"};
  }

  if (workShape === "bounded-transform") {
    if (requestedExecutor && requestedExecutor !== "model") {
      return {code: "executor-not-permitted"};
    }
    return hasEveryValue(bounds, requiredModelBounds)
      ? {executor: "model"}
      : {code: "unbounded-model"};
  }

  if (workShape === "adaptive") {
    if (requestedExecutor && requestedExecutor !== "agent") {
      return {code: "executor-not-permitted"};
    }
    return hasEveryValue(bounds, requiredAgentBounds)
      ? {executor: "agent"}
      : {code: "unbounded-agent"};
  }

  return {code: "unknown-work-shape"};
};

export const validateDevelopment = (development) => {
  if (!development?.planningReference || !development.planningRevision) {
    return "development-not-revision-bound";
  }

  if (development.mode === "run") {
    if (!Array.isArray(development.assignments)) {
      return "development-assignments-missing";
    }
    const groups = Map.groupBy(
      development.assignments,
      ({group}) => group ?? "ungrouped",
    );
    const sharedParallelWorkspace = [...groups.values()].some(
      (assignments) =>
        assignments.length > 1 &&
        duplicateValues(
          assignments.map(({workspaceReference}) => workspaceReference),
        ).length > 0,
    );
    if (sharedParallelWorkspace) return "workspace-isolation-missing";
  }

  if (development.mode === "consolidate") {
    if (development.integrationPerformed) return "consolidation-is-integration";
    if (
      !Array.isArray(development.sourceResults) ||
      development.sourceResults.length === 0 ||
      development.sourceResults.some(
        ({reference, revision}) => !reference || !revision,
      ) ||
      !development.targetWorkspaceReference ||
      !development.targetStartingRevision ||
      !development.resultRevision ||
      !development.conflictPolicy
    ) {
      return "consolidation-contract-incomplete";
    }
  }

  if (
    development.mode === "abandon" &&
    (!development.reason ||
      !development.partialResultReference ||
      !development.cleanupOutcome ||
      development.outcome !== "abandoned")
  ) {
    return "abandonment-contract-incomplete";
  }

  if (
    development.outcome === "partial" &&
    (!development.successfulAssignments?.length ||
      !development.failedAssignments?.length ||
      development.failureBehavior !== "preserve-success")
  ) {
    return "partial-result-not-preserved";
  }

  if (
    development.retry?.enabled &&
    (!development.retry.idempotencyKey ||
      !Number.isInteger(development.retry.retryLimit) ||
      development.retry.retryLimit < 0)
  ) {
    return "unsafe-retry";
  }

  return null;
};

export const validateDeliverable = (deliverable) => {
  if (!deliverable?.providerAvailable) return "explicit-provider-unavailable";
  if (
    !deliverable.providerContext ||
    !deliverable.nativeReference ||
    !deliverable.immutableRevision ||
    !deliverable.intendedTarget ||
    !deliverable.sourceDevelopmentRevisions?.length ||
    !deliverable.publicationStatus
  ) {
    return "deliverable-not-revision-bound";
  }
  return null;
};

export const validateInventory = (inventory, specializations) => {
  if (!inventory?.subjectReference || !inventory.subjectRevision) {
    return "inventory-not-revision-bound";
  }
  if (!specializations.includes(inventory.specialization)) {
    return "unsupported-inventory-specialization";
  }
  if (
    inventory.generator?.executor !== "deterministic" ||
    !inventory.generator.identity ||
    !inventory.generator.version
  ) {
    return "non-deterministic-inventory-generator";
  }
  if (
    !Array.isArray(inventory.components) ||
    inventory.components.some(
      ({identity, version, provenance}) => !identity || !version || !provenance,
    )
  ) {
    return "inventory-component-incomplete";
  }
  if (
    inventory.components.some(({versionOrigin}) => versionOrigin === "guessed")
  ) {
    return "guessed-inventory-fact";
  }
  if (
    inventory.enrichments?.some(
      ({field, origin, authoritative}) =>
        origin === "model" &&
        (field !== "description" || authoritative !== false),
    )
  ) {
    return "model-authoritative-inventory-fact";
  }
  if (inventory.specialization === "agent-environment") {
    const allowedStates = [
      "available",
      "installed",
      "selected",
      "enabled",
      "observed-effective",
      "evidence-producing",
    ];
    if (
      inventory.components.some(
        ({componentKind, effectiveState}) =>
          !componentKind || !allowedStates.includes(effectiveState),
      )
    ) {
      return "agent-environment-state-incomplete";
    }
    if (
      inventory.components.some(
        ({effectiveState, effectiveSelectionEvidence}) =>
          effectiveState === "observed-effective" &&
          !effectiveSelectionEvidence,
      )
    ) {
      return "inventory-effective-state-unproven";
    }
    if (
      inventory.components.some(
        ({componentKind, classification, adoptionModes, authorityZones}) =>
          ["governance-module", "policy-bundle"].includes(componentKind) &&
          (!classification ||
            !adoptionModes?.length ||
            !authorityZones?.length),
      )
    ) {
      return "governance-inventory-placement-incomplete";
    }
  }
  return null;
};

export const validateAssurance = ({assurance, profile}) => {
  if (
    !assurance?.subjectReference ||
    !assurance.subjectRevision ||
    !assurance.criteria?.identity ||
    !assurance.criteria.version
  ) {
    return "assurance-not-revision-bound";
  }
  if (
    ["Review", "QA"].includes(assurance.kind) &&
    assurance.subjectKind !== "DeliverablePublication"
  ) {
    return "deliverable-specific-assurance-required";
  }
  if (assurance.kind === "Review") {
    const independenceCode = checkIndependence({
      required: profile?.independence?.review,
      decider: assurance.assessor,
      author: assurance.subjectAuthor,
      providerEvidence: assurance.independenceEvidenceReference,
      failureCode: "assessor-independence-failed",
    });
    if (independenceCode !== null) return independenceCode;
  }
  if (
    assurance.kind === "QA" &&
    (!assurance.environment?.identity || !assurance.environment.version)
  ) {
    return "qa-environment-missing";
  }
  if (!Array.isArray(assurance.evidence) || assurance.evidence.length === 0) {
    return "assurance-evidence-missing";
  }
  if (assurance.evidence.some(({treatAsInstructions}) => treatAsInstructions)) {
    return "untrusted-evidence-instructions";
  }
  if (
    assurance.evidence.some(
      ({nativeReference, verified}) => !nativeReference || !verified,
    )
  ) {
    return "poisoned-evidence";
  }
  if (assurance.partial && assurance.outcome === "pass") {
    return "partial-evidence-passed";
  }
  return null;
};

export const validateAssuranceBatch = (batch) => {
  const requiredAssessmentFamilies = [
    "security",
    "accessibility",
    "ai",
    "supply-chain",
  ];
  if (
    !batch?.subjectReference ||
    !batch.subjectRevision ||
    !Number.isInteger(batch.maxConcurrency) ||
    batch.maxConcurrency < 2 ||
    !Array.isArray(batch.results)
  ) {
    return "assurance-concurrency-contract-incomplete";
  }
  if (
    batch.results.some(
      ({nativeResultReference, subjectRevision}) =>
        !nativeResultReference || subjectRevision !== batch.subjectRevision,
    ) ||
    duplicateValues(
      batch.results.map(({nativeResultReference}) => nativeResultReference),
    ).length > 0
  ) {
    return "assurance-results-not-independent";
  }
  if (
    !["Review", "QA"].every((kind) =>
      batch.results.some((result) => result.kind === kind),
    ) ||
    !requiredAssessmentFamilies.every((family) =>
      batch.results.some(
        (result) => result.kind === "Assessment" && result.family === family,
      ),
    )
  ) {
    return "assurance-coverage-incomplete";
  }
  return null;
};

export const evaluateFreshness = ({requiredBindings, binding, current}) => {
  if (!Array.isArray(requiredBindings) || requiredBindings.length === 0) {
    return "freshness-binding-missing";
  }
  const changed = requiredBindings.filter(
    (field) =>
      !hasValue(binding?.[field]) || binding[field] !== current?.[field],
  );
  return changed.length > 0 ? `stale-evidence:${changed.join(",")}` : null;
};

export const validateExternalEvidence = (evidence) => {
  if (["github", "gitlab"].includes(evidence?.provider)) {
    if (!/^[0-9a-f]{40}$/u.test(evidence.modulePin ?? "")) {
      return "mutable-ci-module";
    }
  }
  if (
    !evidence?.subjectRevision ||
    !evidence.nativeReference ||
    !evidence.checkIdentity ||
    !evidence.sourceIdentity
  ) {
    return "external-evidence-incomplete";
  }
  if (
    evidence.expectedSourceIdentity &&
    evidence.sourceIdentity !== evidence.expectedSourceIdentity
  ) {
    return "spoofed-check-identity";
  }
  if (
    evidence.mandatory &&
    (!evidence.guaranteed || evidence.governedActorCanBypass)
  ) {
    return "mandatory-control-bypassable";
  }
  if (evidence.outcome !== "success") return "external-evidence-not-successful";
  return null;
};
