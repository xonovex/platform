const requiredProviderOperations = {
  policy: [
    "evaluate",
    "request-evidence",
    "explain",
    "resolve-exception",
    "version",
    "replay",
  ],
  configuration: [
    "inspect",
    "diff",
    "preview",
    "apply",
    "verify",
    "rollback",
    "export",
    "import",
    "detect-drift",
  ],
  evidence: [
    "publish",
    "resolve",
    "correlate",
    "redact",
    "retention",
    "authorize-access",
    "version",
  ],
};

export const expectedVocabulary = {
  executorClasses: ["deterministic", "model", "agent", "human", "external"],
  policyOutcomes: [
    "allow",
    "deny",
    "ask",
    "advise",
    "observe",
    "require-evidence",
    "exception",
    "break-glass",
  ],
  moduleKinds: [
    "script",
    "model-evaluator",
    "agent-launcher",
    "external-job",
    "plugin",
    "skill",
    "mcp-integration",
    "human-task",
  ],
  moduleClassifications: [
    "knowledge-only",
    "advisory",
    "evidence-producing",
    "enforcing",
    "configuration-changing",
    "privileged",
  ],
  authorityZones: [
    "organization-managed",
    "project",
    "user",
    "session-runtime",
    "external",
  ],
  adoptionModes: [
    "workflow-only",
    "governance-only",
    "enablement-only",
    "external-enforcement-only",
    "integrated",
  ],
  provenanceVerificationMethods: [
    "package-metadata",
    "signature",
    "checksum",
    "slsa-attestation",
    "provider-native",
  ],
  semanticEventIntents: [
    "session",
    "prompt",
    "model",
    "tool",
    "capability",
    "result",
    "configuration",
    "context-compaction",
    "subagent",
    "workspace",
    "privileged-operation",
  ],
  capabilityMatrixFields: [
    "event",
    "handler-type",
    "blocking",
    "output-context",
    "ordering-concurrency",
    "managed-configuration",
    "version",
    "limitations",
    "trust-boundary",
  ],
  profileFacets: [
    "lifecycle",
    "governance",
    "executor",
    "enforcement",
    "data",
    "telemetry",
    "distribution",
  ],
  onboardingStages: [
    "discover",
    "assess",
    "recommend",
    "preview",
    "approve",
    "apply",
    "verify",
    "rollback",
    "drift",
    "upgrade",
    "remove",
  ],
};

export const validateVocabulary = (fixture) => {
  const failures = Object.entries(expectedVocabulary).flatMap(
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
    throw new Error(`governance vocabulary failure: ${failures.join(",")}`);
  }
};

export const validateExecutor = (executor) => {
  const requiredValues = [
    executor.identity,
    executor.version,
    executor.purpose,
    executor.executorClass,
    executor.resultContract,
    executor.sideEffects,
    executor.inputValidation,
    executor.outputValidation,
    executor.evidenceOrigin,
    executor.timeoutSeconds,
    executor.retryLimit,
    executor.concurrency,
    executor.ordering,
    executor.cancellation,
    executor.killBehavior,
    executor.failureBehavior,
  ];
  const requiredArrays = [
    executor.inputReferences,
    executor.toolScope,
    executor.filesystemScope,
    executor.networkScope,
    executor.secretScope,
    executor.modelScope,
    executor.providerScope,
    executor.dataScope,
  ];
  const invalidBudget =
    !Number.isFinite(executor.timeoutSeconds) ||
    executor.timeoutSeconds <= 0 ||
    !Number.isInteger(executor.retryLimit) ||
    executor.retryLimit < 0;

  if (
    requiredValues.some((value) => value === null || value === undefined) ||
    requiredArrays.some((value) => !Array.isArray(value)) ||
    invalidBudget
  ) {
    return "executor-contract-incomplete";
  }
  return executor.executorClass !== "deterministic" &&
    executor.authoritativeInspectionAvailable &&
    executor.outputTreatedAsAuthoritative
    ? "authoritative-evidence-replaced"
    : null;
};

export const validateEnforcement = ({mandatory, capability, references}) => {
  if (
    mandatory &&
    (capability?.support !== "supported" || capability?.blocking !== true)
  ) {
    return "unsupported-enforcement";
  }
  const decisionReference = references?.decision;
  const enforcementReference = references?.enforcement;
  const evidenceReference = references?.evidence;
  const uniqueReferences = new Set([
    decisionReference,
    enforcementReference,
    evidenceReference,
  ]);
  return !decisionReference ||
    !enforcementReference ||
    !evidenceReference ||
    uniqueReferences.size !== 3
    ? "decision-enforcement-conflated"
    : null;
};

export const validateComposition = ({modules, conflictResolution}) => {
  const ids = new Set(modules.map(({id}) => id));
  const hasConflict = modules.some(({conflictsWith}) =>
    conflictsWith.some((conflict) => ids.has(conflict)),
  );
  return hasConflict && !conflictResolution ? "module-conflict" : null;
};

export const validateOrdering = ({requiresSerial, adapterOrdering}) =>
  requiresSerial && adapterOrdering !== "serial"
    ? "ordering-unguaranteed"
    : null;

export const validateAgent = (agent) => {
  const requiredValues = [
    agent.purpose,
    agent.maxDepth,
    agent.modelProvider,
    agent.timeBudgetSeconds,
    agent.tokenBudget,
    agent.costBudget,
    agent.resultContract,
    agent.outputValidation,
    agent.evidence,
    agent.cancellation,
    agent.failureBehavior,
  ];
  const requiredArrays = [
    agent.toolScope,
    agent.filesystemScope,
    agent.networkScope,
    agent.secretScope,
    agent.dataScope,
  ];
  const invalidBudget =
    !Number.isInteger(agent.maxDepth) ||
    agent.maxDepth < 0 ||
    !Number.isFinite(agent.timeBudgetSeconds) ||
    agent.timeBudgetSeconds <= 0 ||
    !Number.isFinite(agent.tokenBudget) ||
    agent.tokenBudget <= 0 ||
    !Number.isFinite(agent.costBudget) ||
    agent.costBudget < 0;
  return requiredValues.some(
    (value) => value === null || value === undefined,
  ) ||
    requiredArrays.some((value) => !Array.isArray(value)) ||
    invalidBudget ||
    !agent.authorityAttenuated ||
    !agent.observable ||
    !agent.killSwitch
    ? "unbounded-agent"
    : null;
};

export const validatePolicy = ({mandatory, fresh}) =>
  mandatory && !fresh ? "stale-policy" : null;

export const validateException = (exception, evaluationTime) => {
  const start = Date.parse(exception.start);
  const expiry = Date.parse(exception.expiry);
  const complete =
    exception.owner &&
    exception.scope &&
    exception.control &&
    exception.authorizedApprover &&
    exception.rationale &&
    exception.compensatingControls?.length > 0 &&
    exception.evidence?.length > 0 &&
    exception.affectedSubjects?.length > 0 &&
    Number.isFinite(start) &&
    Number.isFinite(expiry) &&
    start < expiry &&
    exception.reviewRequired;
  if (!complete) return "incomplete-exception";
  return expiry <= evaluationTime ? "exception-expired" : null;
};

export const validateBreakGlass = (breakGlass, evaluationTime) => {
  const exceptionCode = validateException(breakGlass, evaluationTime);
  if (exceptionCode) return exceptionCode;
  const complete =
    breakGlass.emergencyReason &&
    breakGlass.timeLimitedAccess &&
    breakGlass.explicitInvocation &&
    breakGlass.authoritativeAccessEvidence?.length > 0 &&
    breakGlass.notificationRequired &&
    breakGlass.containmentPlan &&
    breakGlass.revocationRequired &&
    breakGlass.postEventReviewRequired;
  return complete ? null : "incomplete-break-glass";
};

export const validateNativeProvider = ({nativeReference, requiresFile}) =>
  nativeReference && !requiresFile ? null : "provider-contract-invalid";

export const validateProfile = ({
  facets,
  strengthening,
  weakensMandatoryControl,
  authorizedException,
  adequateEnforcement,
}) => {
  const missingFacets = expectedVocabulary.profileFacets.filter(
    (facet) => !facets?.includes(facet),
  );
  if (missingFacets.length > 0 || strengthening !== "additive") {
    return "profile-composition-incomplete";
  }
  if (weakensMandatoryControl && !authorizedException)
    return "unauthorized-weakening";
  return adequateEnforcement ? null : "no-enforcement-guarantee";
};

export const validateProviderContract = (provider) => {
  const expectedOperations = requiredProviderOperations[provider.providerKind];
  if (!expectedOperations) return "unknown-provider-kind";
  if (!provider.storageNeutral || !provider.nativeReferencesOpaque) {
    return "storage-coupled-provider";
  }
  const missingOperations = expectedOperations.filter(
    (operation) => !provider.operations?.includes(operation),
  );
  if (missingOperations.length > 0) {
    return `missing-provider-operation:${missingOperations.join(",")}`;
  }
  if (!provider.versioned || !provider.failureBehavior) {
    return "provider-lifecycle-unspecified";
  }
  if (!provider.reconstructableAfterRestart) {
    return "provider-not-reconstructable";
  }
  if (
    provider.providerKind === "policy" &&
    (!provider.deterministic || !provider.historicalReplay)
  ) {
    return "policy-replay-unsupported";
  }
  if (
    provider.providerKind === "configuration" &&
    (!provider.previewBeforeApply ||
      !provider.idempotentApply ||
      !provider.verifiableRollback ||
      !provider.driftDetection)
  ) {
    return "configuration-transaction-incomplete";
  }
  if (
    provider.providerKind === "evidence" &&
    (!provider.dataMinimization ||
      !provider.redaction ||
      !provider.retention ||
      !provider.accessControl ||
      !provider.correlation)
  ) {
    return "evidence-governance-incomplete";
  }
  return null;
};

const decisionKey = (decision = {}) => {
  const {
    policyVersion,
    subject,
    outcome,
    reasons,
    evidenceRequests,
    explanation,
    exception,
  } = decision;
  return JSON.stringify({
    policyVersion,
    subject,
    outcome,
    reasons,
    evidenceRequests,
    explanation,
    exception,
  });
};

export const validatePolicyParity = (policyCase) => {
  if (policyCase.implementations.length < 2) {
    return "policy-portability-unproven";
  }
  const semanticDecisions = policyCase.implementations.map(
    ({semanticDecision}) => semanticDecision,
  );
  const nativeOutcomes = policyCase.implementations.map(
    ({implementation, nativeOutput}) =>
      implementation === "opa"
        ? nativeOutput.result?.[0]?.expressions?.[0]?.value
        : nativeOutput.outcome,
  );
  if (
    nativeOutcomes.some(
      (outcome, index) => outcome !== semanticDecisions[index].outcome,
    )
  ) {
    return "policy-adapter-output-mismatch";
  }
  if (semanticDecisions.some((decision) => !decision.decisionReference)) {
    return "policy-decision-reference-missing";
  }
  if (
    policyCase.implementations.some(
      ({historicalReplay, replayDecision}) =>
        !historicalReplay ||
        decisionKey(replayDecision) !== decisionKey(policyCase.expected),
    )
  ) {
    return "policy-replay-diverged";
  }
  return semanticDecisions.every(
    (decision) => decisionKey(decision) === decisionKey(policyCase.expected),
  )
    ? null
    : "policy-provider-divergence";
};

export const validateEvidenceCase = (evidence) => {
  if (!evidence.nativeReference || !evidence.correlationReference) {
    return "evidence-reference-missing";
  }
  if (evidence.capturesSensitiveContent && !evidence.contentCaptureAuthorized) {
    return "sensitive-content-unauthorized";
  }
  if (
    !evidence.redacted ||
    !evidence.retentionDefined ||
    !evidence.accessControlled
  ) {
    return "evidence-governance-incomplete";
  }
  return null;
};

export const validateConfigurationCase = (configuration) => {
  if (
    configuration.applied &&
    (!configuration.previewed || !configuration.authorized)
  ) {
    return "configuration-applied-without-authorization";
  }
  if (configuration.applied && !configuration.verified) {
    return "configuration-unverified";
  }
  if (
    configuration.partialApplication &&
    configuration.rollbackStatus !== "succeeded"
  ) {
    return "rollback-failure";
  }
  if (configuration.driftDetected && !configuration.driftReported) {
    return "drift-unreported";
  }
  if (configuration.retryCount > 0 && !configuration.idempotencyKey) {
    return "duplicate-risk-uncontrolled";
  }
  return null;
};

const isRecord = (value) =>
  value !== null && typeof value === "object" && !Array.isArray(value);

export const mergeFixture = (base, overrides) =>
  Object.fromEntries(
    Object.keys({...base, ...overrides}).map((key) => [
      key,
      isRecord(base[key]) && isRecord(overrides[key])
        ? mergeFixture(base[key], overrides[key])
        : structuredClone(
            Object.hasOwn(overrides, key) ? overrides[key] : base[key],
          ),
    ]),
  );

export const validateModule = (module, requiredFields) => {
  if (!module.source || !module.provenance) return "missing-provenance";
  const missingFields = requiredFields.filter(
    (field) => !Object.hasOwn(module, field),
  );
  if (missingFields.length > 0) {
    return `module-field-missing:${missingFields.join(",")}`;
  }
  if (!module.provenance.versionPinned) return "moving-version";
  if (
    !Array.isArray(module.provenance.verification) ||
    module.provenance.verification.length === 0
  ) {
    return "missing-provenance";
  }
  if (
    module.provenance.expectedDigest &&
    module.provenance.expectedDigest !== module.provenance.observedDigest
  ) {
    return "tampered-module";
  }
  if (!module.provenance.verification.some(({verified}) => verified)) {
    return "provenance-unverified";
  }
  if (
    module.provenance.verification.some(
      ({method}) =>
        !expectedVocabulary.provenanceVerificationMethods.includes(method),
    )
  ) {
    return "provenance-method-unknown";
  }
  if (
    !Array.isArray(module.permissions?.declared) ||
    !Array.isArray(module.permissions?.approved) ||
    !Array.isArray(module.permissions?.observed)
  ) {
    return "module-field-invalid:permissions";
  }
  if (
    !Array.isArray(module.adoptionModes) ||
    !Array.isArray(module.authorityZones)
  ) {
    return "module-placement-invalid";
  }
  if (!isRecord(module.trust)) return "module-field-invalid:trust";
  const permissionExpansion = module.permissions.observed.some(
    (permission) =>
      !module.permissions.declared.includes(permission) ||
      !module.permissions.approved.includes(permission),
  );
  if (permissionExpansion) return "unexpected-permissions";
  if (
    module.authorityZones.some((zone) => ["project", "user"].includes(zone)) &&
    !module.trust.reviewed
  ) {
    return "trust-review-required";
  }
  if (module.authorityZones.includes("user") && !module.trust.userConsent) {
    return "user-consent-required";
  }
  if (
    module.authorityZones.includes("organization-managed") &&
    !module.trust.organizationChangeControlled
  ) {
    return "organization-provenance-uncontrolled";
  }
  if (
    !expectedVocabulary.moduleClassifications.includes(module.classification)
  ) {
    return "module-classification-invalid";
  }
  if (
    module.adoptionModes.some(
      (mode) => !expectedVocabulary.adoptionModes.includes(mode),
    ) ||
    module.authorityZones.some(
      (zone) => !expectedVocabulary.authorityZones.includes(zone),
    )
  ) {
    return "module-placement-invalid";
  }
  if (!Number.isFinite(module.timeoutMs) || module.timeoutMs <= 0) {
    return "timeout-unspecified";
  }
  if (module.retryLimit > 0 && !module.idempotent) {
    return "non-idempotent-retry";
  }
  if (module.concurrency !== "serial" && !module.reentrant) {
    return "reentrancy-unsafe";
  }
  if (
    !["fail-closed", "fail-visible", "advisory"].includes(module.failureMode)
  ) {
    return "failure-mode-invalid";
  }
  if (
    module.concurrency !== "serial" &&
    module.sideEffects !== "none" &&
    !module.duplicateSafe
  ) {
    return "duplicate-execution-unsafe";
  }
  if (module.sideEffects !== "none" && !module.rollback?.supported) {
    return "rollback-unavailable";
  }
  if (module.rollback?.status === "failed" || !module.rollback?.verified) {
    return "rollback-failure";
  }
  return null;
};
