import {expectedVocabulary} from "./conformance-helpers.mjs";

const allowedPolicyDomains = [
  "security",
  "privacy",
  "accessibility",
  "reliability",
  "ai-governance",
  "supply-chain",
  "data",
  "cost",
  "regulated",
];

const allowedDecisionProviders = ["deterministic", "provider-native", "opa"];

const allowedFailureModes = ["fail-closed", "fail-visible", "advisory"];

const allowedModuleClassifications = expectedVocabulary.moduleClassifications;

const requiredCatalogFields = [
  "id",
  "version",
  "owner",
  "classification",
  "adoptionModes",
  "authorityZones",
  "compatibility",
  "trust",
  "permissions",
  "conflictsWith",
  "presets",
  "lifecycleStatus",
  "usage",
  "deprecation",
];

const allowedLearningSources = [
  "lifecycle",
  "onboarding",
  "policy-denial",
  "incident",
  "exception",
  "drift",
  "module-failure",
];

export const expectedOperationalVocabulary = {
  policyDomains: allowedPolicyDomains,
  decisionProviders: allowedDecisionProviders,
  failureModes: allowedFailureModes,
  moduleClassifications: allowedModuleClassifications,
  learningSources: allowedLearningSources,
  driftDimensions: [
    "harness-configuration",
    "ci-policy",
    "module-version",
    "provider-capability",
    "profile-control-version",
    "managed-setting",
  ],
};

export const validateOperationalVocabulary = (fixture) => {
  const failures = Object.entries(expectedOperationalVocabulary).flatMap(
    ([key, expected]) => {
      const actual = fixture[key];
      const matches =
        Array.isArray(actual) &&
        expected.length === actual.length &&
        expected.every((value) => actual.includes(value));
      return matches ? [] : [key];
    },
  );
  if (failures.length > 0) {
    throw new Error(
      `governance operations vocabulary failure: ${failures.join(",")}`,
    );
  }
};

export const validatePolicyBundle = (bundle) => {
  if (!allowedPolicyDomains.includes(bundle.domain)) {
    return "policy-domain-unknown";
  }
  if (!bundle.versionPinned || !bundle.owner || !bundle.applicabilityResolved) {
    return "policy-bundle-unresolved";
  }
  if (!allowedDecisionProviders.includes(bundle.decisionProvider)) {
    return "policy-provider-unsupported";
  }
  if (!allowedFailureModes.includes(bundle.failureMode)) {
    return "failure-mode-invalid";
  }
  if (!bundle.enginePortable) return "policy-engine-mandatory";
  if (bundle.mandatory && !bundle.enforcement?.adequate) {
    return "missing-enforcement";
  }
  if (
    ["advisory", "telemetry"].includes(bundle.controlKind) &&
    bundle.failureMode === "fail-closed"
  ) {
    return "advisory-overblocking";
  }
  const {decision, enforcement, evidence} = bundle.references ?? {};
  return decision && enforcement && evidence
    ? null
    : "policy-evidence-incomplete";
};

export const validateDataGovernance = (dataCase) => {
  if (
    !dataCase.classification ||
    !dataCase.purpose ||
    !dataCase.minimized ||
    !dataCase.retentionDefined ||
    !dataCase.residencyDefined ||
    !dataCase.accessControlled
  ) {
    return "data-governance-incomplete";
  }
  if (dataCase.sensitiveContentCaptured && !dataCase.captureAuthorized) {
    return "sensitive-content-unauthorized";
  }
  if (dataCase.sensitiveContentCaptured && !dataCase.redacted) {
    return "sensitive-content-unredacted";
  }
  if (dataCase.externalTransfer && !dataCase.transferAuthorized) {
    return "external-transfer-unauthorized";
  }
  if (!dataCase.modelRouteAllowed) return "model-route-disallowed";
  return dataCase.telemetrySchemaPinned ? null : "telemetry-schema-unpinned";
};

export const validateCatalogEntry = (entry, selectedIds = []) => {
  const missingFields = requiredCatalogFields.filter(
    (field) => !Object.hasOwn(entry, field),
  );
  if (missingFields.length > 0) {
    return `catalog-field-missing:${missingFields.join(",")}`;
  }
  if (!allowedModuleClassifications.includes(entry.classification)) {
    return "module-classification-invalid";
  }
  if (
    !entry.compatibility.tested ||
    !entry.trust.reviewed ||
    !entry.permissions.declared
  ) {
    return "catalog-assurance-incomplete";
  }
  if (entry.conflictsWith.some((id) => selectedIds.includes(id))) {
    return "catalog-conflict";
  }
  if (entry.lifecycleStatus === "retired" && entry.usage.selected) {
    return "retired-module-selected";
  }
  if (
    entry.lifecycleStatus === "deprecated" &&
    entry.usage.selected &&
    !entry.deprecation.replacement
  ) {
    return "deprecation-path-missing";
  }
  return null;
};

export const validateDriftCase = (drift) => {
  if (drift.intendedVersion === drift.observedVersion) return null;
  if (!drift.reported || !drift.nativeReference) return "drift-unreported";
  if (
    drift.mandatoryGuaranteeInvalidated &&
    !["fail-closed", "fail-visible"].includes(drift.response)
  ) {
    return "mandatory-drift-silent";
  }
  return null;
};

export const validateUpdateCase = (update) => {
  if (!update.candidatePinned || !update.compatibilityVerified) {
    return "update-target-unverified";
  }
  if (
    !update.previewed ||
    !update.authorized ||
    !update.canary?.representative ||
    !update.canary?.successCriteria ||
    !update.canary?.abortCriteria
  ) {
    return "update-gate-incomplete";
  }
  if (!update.rollback?.targetPinned || !update.rollback?.verified) {
    return "rollback-unavailable";
  }
  if (!update.emergencyDisable?.independent) {
    return "emergency-disable-unavailable";
  }
  return update.coverageChecked ? null : "control-coverage-unverified";
};

export const validateLearningCandidate = (candidate) => {
  if (!allowedLearningSources.includes(candidate.source)) {
    return "learning-source-unknown";
  }
  if (
    !candidate.evidence ||
    !candidate.scope ||
    !candidate.owner ||
    !candidate.target
  ) {
    return "learning-candidate-incomplete";
  }
  if (candidate.autoPromoted) return "learning-auto-promoted";
  return candidate.reviewed &&
    candidate.versioned &&
    candidate.conflictChecked &&
    candidate.measurable &&
    candidate.reversible
    ? null
    : "learning-promotion-ungoverned";
};

export const validateMetric = (metric) => {
  if (!metric.purpose || !metric.audience || !metric.aggregated) {
    return "metric-purpose-incomplete";
  }
  if (metric.individualRanking || metric.automaticPunishment) {
    return "surveillance-metric";
  }
  if (
    !metric.accessControlled ||
    !metric.retentionDefined ||
    !metric.interpretationLimits
  ) {
    return "metric-governance-incomplete";
  }
  return metric.counterMetrics?.length > 0 && metric.qualitativeReview
    ? null
    : "metric-gaming-uncontrolled";
};

export const validateOperationalOwnership = (operation) => {
  if (!operation.owner || !operation.supportStatus || !operation.escalation) {
    return "operational-owner-missing";
  }
  if (!operation.incidentPlan || !operation.retirementPlan) {
    return "operational-lifecycle-incomplete";
  }
  return operation.evidenceRetentionDefined
    ? null
    : "operational-evidence-undefined";
};

const crossPlatformValidators = {
  conflict: (testCase) => testCase.resolved || "module-conflict",
  concurrency: (testCase) =>
    (testCase.reentrant && testCase.duplicateSafe) || "concurrency-unsafe",
  trust: (testCase) =>
    (testCase.provenanceVerified && testCase.permissionsApproved) ||
    "trust-unverified",
  "data-leakage": (testCase) =>
    (!testCase.sensitiveContentLogged && testCase.redactionVerified) ||
    "sensitive-content-leaked",
  "stale-version": (testCase) =>
    (testCase.driftReported && testCase.failureVisible) ||
    "stale-version-silent",
  "missing-enforcement": (testCase) =>
    !testCase.mandatory ||
    testCase.adequateEnforcement ||
    "missing-enforcement",
  "exception-abuse": (testCase) =>
    (testCase.scoped &&
      testCase.owned &&
      testCase.unexpired &&
      testCase.reviewed &&
      !testCase.permanentDefault) ||
    "exception-abuse",
  "telemetry-outage": (testCase) =>
    (testCase.outageVisible &&
      testCase.failureMode !== "fail-closed" &&
      !testCase.sensitiveFallback) ||
    "telemetry-outage-mishandled",
};

export const validateCrossPlatformCase = (testCase) => {
  const validate = crossPlatformValidators[testCase.risk];
  if (!validate) return "cross-platform-risk-unknown";
  const result = validate(testCase);
  return result === true ? null : result;
};
