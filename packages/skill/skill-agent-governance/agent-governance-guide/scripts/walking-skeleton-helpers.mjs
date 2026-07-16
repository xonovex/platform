import {
  findCapability,
  indexMatrices,
  validateMatrix,
} from "./harness-conformance-helpers.mjs";

const requiredAdoptionModes = [
  "governance-only",
  "workflow-only",
  "integrated",
];
const requiredLifecycleEvents = [
  "discover",
  "assess",
  "recommend",
  "preview",
  "authorize",
  "apply",
  "verify",
  "record",
  "operate",
];
const requiredRecommendationModules = [
  "context-injection",
  "protected-path",
  "tool-policy",
  "validation",
  "audit",
  "github-required-check",
];
const requiredProfileFacets = [
  "governance:",
  "execution:",
  "enforcement:",
  "data:",
  "telemetry:",
];
const requiredNegativeKinds = [
  "module-trust",
  "capability",
  "policy-outage",
  "duplicate-event",
  "agent-recursion",
  "exception-expiry",
];

const isNonEmptyString = (value) =>
  typeof value === "string" && value.length > 0;

const hasSameMembers = (left, right) =>
  left.length === right.length &&
  new Set(left).size === left.length &&
  new Set(right).size === right.length &&
  left.every((value) => right.includes(value));

const hasSameOrder = (left, right) =>
  left.length === right.length &&
  left.every((value, index) => value === right[index]);

const isOpaqueReference = (value) =>
  isNonEmptyString(value) && value.startsWith("opaque:");

const indexById = (items) => new Map(items.map((item) => [item.id, item]));

const validateContractReferences = (fixtures, scenario, fail) => {
  const references = scenario.contractReferences;
  const matrices = indexMatrices(fixtures.harness);
  const matrix = matrices.get(references.harness.platform);
  if (
    !matrix ||
    matrix.matrixVersion !== references.harness.matrixVersion ||
    validateMatrix(matrix, fixtures.harness) !== null
  ) {
    fail("contract:harness-matrix-unresolved");
  }

  const platformFixtures = indexById(fixtures.external.platformFixtures);
  const hostedCi = platformFixtures.get(references.hostedCi.fixtureId);
  if (
    !hostedCi?.expectedValid ||
    hostedCi.platform !== references.hostedCi.platform
  ) {
    fail("contract:hosted-ci-fixture-unresolved");
  }
  if (
    !platformFixtures.get(references.operatorPolicyFixtureId)?.expectedValid
  ) {
    fail("contract:operator-policy-fixture-unresolved");
  }

  const externalCompositions = indexById(fixtures.external.compositionCases);
  if (
    !externalCompositions.get(references.externalCompositionId)?.expectedValid
  ) {
    fail("contract:external-composition-unresolved");
  }

  if (references.enterpriseComposition.fixtureId !== "mixedStack") {
    fail("contract:enterprise-composition-unresolved");
  }
  if (
    !indexById(fixtures.enterprise.federationCases).get(
      references.enterpriseComposition.federationCaseId,
    )?.expectedValid
  ) {
    fail("contract:enterprise-federation-unresolved");
  }
  if (
    !indexById(fixtures.enterprise.telemetryCases).get(
      references.enterpriseComposition.telemetryCaseId,
    )?.expectedValid
  ) {
    fail("contract:enterprise-telemetry-unresolved");
  }

  const evidenceProvider = indexById(fixtures.governance.cases).get(
    references.evidenceProviderFixtureId,
  );
  if (
    !evidenceProvider?.expectedValid ||
    evidenceProvider.kind !== "provider" ||
    evidenceProvider.requiresFile
  ) {
    fail("contract:non-file-provider-unresolved");
  }
};

const validateDiscovery = (fixtures, scenario, fail) => {
  const matrices = indexMatrices(fixtures.harness);
  const discovered = scenario.discovery;
  const harness = discovered.harness;
  const matrix = matrices.get(harness.platform);
  if (
    !matrix ||
    !isNonEmptyString(harness.runtimeVersion) ||
    harness.probeCommand !== matrix.runtimeProbe.command ||
    harness.probeStatus !== "probed" ||
    harness.matrixVersion !== matrix.matrixVersion ||
    harness.surface !== matrix.surface ||
    !isOpaqueReference(harness.diagnosticsReference)
  ) {
    fail("discovery:harness-facts-incomplete");
    return;
  }

  for (const capabilityId of harness.supportedCapabilities) {
    const capability = findCapability(matrices, harness.platform, capabilityId);
    if (capability?.support !== "supported" || !capability.executes) {
      fail(`discovery:unsupported-capability:${capabilityId}`);
    }
    if (!harness.handlerTypes.includes(capability?.handlerType)) {
      fail(`discovery:handler-type-missing:${capabilityId}`);
    }
  }
  if (
    !hasSameMembers(harness.configurationScopes, matrix.configuration.scopes)
  ) {
    fail("discovery:configuration-scopes-incomplete");
  }

  const hostedCi = indexById(fixtures.external.platformFixtures).get(
    scenario.contractReferences.hostedCi.fixtureId,
  );
  const discoveredCi = discovered.ciControls.find(
    ({platform}) => platform === hostedCi?.platform,
  );
  if (
    !discoveredCi ||
    discoveredCi.requiredCheck !== hostedCi.requiredCheck ||
    discoveredCi.rulesetActive !== hostedCi.rulesetActive ||
    !hasSameMembers(discoveredCi.bypassActors, hostedCi.bypassActors)
  ) {
    fail("discovery:ci-controls-incomplete");
  }

  const operator = indexById(fixtures.external.platformFixtures).get(
    scenario.contractReferences.operatorPolicyFixtureId,
  );
  if (
    discovered.operatorControls.agentPolicyCount !==
      operator?.agentPolicyCount ||
    discovered.operatorControls.policyLookupFailure !==
      operator?.policyLookupFailure ||
    discovered.operatorControls.runtimeClass !== operator?.runtimeClass
  ) {
    fail("discovery:operator-controls-incomplete");
  }

  if (
    !Array.isArray(discovered.installedModules) ||
    Object.values(discovered.providerCapabilities).some(
      (capabilities) =>
        !Array.isArray(capabilities) || capabilities.length === 0,
    ) ||
    requiredProfileFacets.some(
      (facet) =>
        !discovered.selectedProfiles.some((profile) =>
          profile.startsWith(facet),
        ),
    )
  ) {
    fail("discovery:composition-facts-incomplete");
  }
};

const validateAdoptionPaths = (fixtures, scenario, fail) => {
  const paths = indexById(
    scenario.adoptionPaths.map((path) => ({...path, id: path.mode})),
  );
  if (
    !hasSameMembers(
      scenario.adoptionPaths.map(({mode}) => mode),
      requiredAdoptionModes,
    ) ||
    requiredAdoptionModes.some(
      (mode) => !fixtures.governance.adoptionModes.includes(mode),
    )
  ) {
    fail("adoption:required-path-missing");
    return;
  }

  const governanceOnly = paths.get("governance-only");
  const workflowOnly = paths.get("workflow-only");
  const integrated = paths.get("integrated");
  if (
    !governanceOnly?.ordinaryAgentTask ||
    governanceOnly.workflowCapabilitySelected ||
    governanceOnly.governanceModules.length === 0
  ) {
    fail("adoption:governance-only-coupled");
  }
  if (
    workflowOnly?.ordinaryAgentTask ||
    !workflowOnly?.workflowCapabilitySelected ||
    workflowOnly.governanceModules.length !== 0 ||
    workflowOnly.resultReferences.length === 0
  ) {
    fail("adoption:workflow-only-coupled");
  }
  if (
    !integrated?.ordinaryAgentTask ||
    !integrated.workflowCapabilitySelected ||
    integrated.governanceModules.length === 0
  ) {
    fail("adoption:integrated-path-incomplete");
  }
  for (const path of scenario.adoptionPaths) {
    if (!path.resultReferences.every(isOpaqueReference)) {
      fail(`adoption:non-opaque-result:${path.mode}`);
    }
  }
};

const validateRecommendation = (fixtures, scenario, fail) => {
  const moduleIds = scenario.recommendation.modules.map(({id}) => id);
  if (
    !hasSameMembers(moduleIds, requiredRecommendationModules) ||
    scenario.recommendation.conflicts.length > 0 ||
    scenario.recommendation.unsupportedRequirements.length > 0
  ) {
    fail("recommendation:minimal-composition-invalid");
  }

  const deterministicTemplates = new Set(
    fixtures.templates.deterministic.map(({id}) => id),
  );
  const matrices = indexMatrices(fixtures.harness);
  for (const module of scenario.recommendation.modules) {
    if (
      module.id !== "github-required-check" &&
      !deterministicTemplates.has(module.id)
    ) {
      fail(`recommendation:unknown-template:${module.id}`);
    }
    if (
      module.id !== "github-required-check" &&
      !findCapability(
        matrices,
        scenario.contractReferences.harness.platform,
        module.nativeCapability,
      )
    ) {
      fail(`recommendation:unknown-capability:${module.id}`);
    }
  }
};

const validatePreviewAndApplication = (fixtures, scenario, fail) => {
  const preview = scenario.preview;
  if (
    !isOpaqueReference(preview.reference) ||
    !isNonEmptyString(preview.digest) ||
    preview.nativeChanges.length === 0 ||
    !preview.nativeChanges.some(({owner}) => owner === "claude-code") ||
    !preview.nativeChanges.some(({owner}) => owner === "github") ||
    preview.nativeChanges.some(
      ({beforeDigest, afterDigest}) =>
        !beforeDigest || !afterDigest || beforeDigest === afterDigest,
    ) ||
    Object.values(preview.requestedPermissions).some(
      (permissions) => !Array.isArray(permissions),
    ) ||
    preview.requestedPermissions.secrets.length > 0 ||
    preview.moduleTrust.authorityZone !== "organization-managed" ||
    !preview.moduleTrust.source ||
    !/^[0-9a-f]{40}$/u.test(preview.moduleTrust.revision) ||
    !preview.moduleTrust.digest ||
    !preview.nativeChanges.some(
      ({afterDigest}) => afterDigest === preview.moduleTrust.digest,
    ) ||
    !isOpaqueReference(preview.moduleTrust.provenanceReference) ||
    !preview.moduleTrust.verified ||
    preview.moduleTrust.permissionExpansion ||
    preview.dataFlow.some(
      ({destination, categories, contentCapture, redacted}) =>
        !preview.requestedPermissions.network.includes(destination) ||
        !Array.isArray(categories) ||
        categories.length === 0 ||
        contentCapture ||
        !redacted,
    ) ||
    !preview.modelUse.purpose ||
    preview.modelUse.timeBudgetSeconds <= 0 ||
    preview.modelUse.tokenBudget <= 0 ||
    preview.modelUse.costBudget < 0 ||
    preview.modelUse.failureMode !== "fail-visible" ||
    preview.failureBehavior.mandatoryPolicy !== "fail-closed" ||
    preview.failureBehavior.advisoryEvaluator !== "fail-visible" ||
    preview.failureBehavior.evidencePublication !== "fail-visible" ||
    preview.verification.length === 0 ||
    !preview.rollback.available ||
    !preview.rollback.retainsEvidence
  ) {
    fail("onboarding:preview-incomplete");
  }

  const authorization = scenario.authorization;
  if (
    !authorization.authorized ||
    authorization.previewReference !== preview.reference ||
    authorization.previewDigest !== preview.digest ||
    !isOpaqueReference(authorization.reference) ||
    !authorization.actor ||
    authorization.authority !== scenario.subject.authorityZone ||
    !Number.isFinite(Date.parse(authorization.expiresAt))
  ) {
    fail("onboarding:authorization-not-bound-to-preview");
  }
  if (!hasSameOrder(scenario.lifecycleEvents, requiredLifecycleEvents)) {
    fail("onboarding:lifecycle-incomplete");
  }
  const authorizationIndex = scenario.lifecycleEvents.indexOf("authorize");
  const applicationIndex = scenario.lifecycleEvents.indexOf("apply");
  if (authorizationIndex < 0 || applicationIndex <= authorizationIndex) {
    fail("onboarding:apply-before-authorization");
  }

  const application = scenario.application;
  const recommendedModules = scenario.recommendation.modules.map(({id}) => id);
  if (
    !application.applied ||
    !application.verified ||
    !application.idempotencyKey ||
    application.authorizationReference !== authorization.reference ||
    !hasSameMembers(application.installedModules, recommendedModules) ||
    application.nativeDiagnostics.length === 0 ||
    !isOpaqueReference(application.applyReference) ||
    !isOpaqueReference(application.verificationReference)
  ) {
    fail("onboarding:application-unverified");
  }

  const matrices = indexMatrices(fixtures.harness);
  for (const diagnostic of application.nativeDiagnostics) {
    const capability = findCapability(
      matrices,
      scenario.contractReferences.harness.platform,
      diagnostic.capability,
    );
    if (
      !capability ||
      capability.probe !== diagnostic.probe ||
      diagnostic.outcome !== "pass" ||
      !isOpaqueReference(diagnostic.reference)
    ) {
      fail(`onboarding:native-diagnostic-failed:${diagnostic.capability}`);
    }
  }
};

const validateExecution = (fixtures, scenario, fail) => {
  const probes = indexById(scenario.policyProbes);
  const allowed = probes.get("permitted-read");
  const denied = probes.get("denied-protected-write");
  if (
    !hasSameMembers(
      scenario.policyProbes.map(({id}) => id),
      ["permitted-read", "denied-protected-write"],
    )
  ) {
    fail("execution:policy-probes-incomplete");
  }
  for (const probe of [allowed, denied]) {
    if (
      !probe?.reason ||
      !isOpaqueReference(probe.decisionReference) ||
      !isOpaqueReference(probe.enforcementReference) ||
      probe.decisionReference === probe.enforcementReference
    ) {
      fail(`execution:policy-evidence-incomplete:${probe?.id ?? "missing"}`);
    }
  }
  if (allowed?.outcome !== "allow" || denied?.outcome !== "deny") {
    fail("execution:deterministic-policy-outcomes-invalid");
  }

  const review = scenario.advisoryReview;
  const template = fixtures.templates.modelEvaluator;
  const outputKeys = Object.keys(review.output);
  if (
    review.moduleId !== template.id ||
    review.executorClass !== "model" ||
    !review.advisory ||
    review.authoritative ||
    review.timeBudgetSeconds > template.timeBudgetSeconds ||
    review.tokenBudget > template.tokenBudget ||
    review.costBudget > template.costBudget ||
    review.retryLimit > template.retryLimit ||
    review.toolScope.length > 0 ||
    review.filesystemScope.length > 0 ||
    review.secretScope.length > 0 ||
    !review.networkScope.every((scope) =>
      template.networkScope.includes(scope),
    ) ||
    !hasSameMembers(outputKeys, template.outputSchema.required) ||
    !Array.isArray(review.output.reasons) ||
    !Array.isArray(review.output.evidenceRequests) ||
    !review.output.outcome ||
    !isOpaqueReference(review.resultReference) ||
    review.failureMode !== "fail-visible"
  ) {
    fail("execution:model-evaluator-unbounded");
  }
};

const validateEvidence = (fixtures, scenario, fail) => {
  const evidence = scenario.evidence;
  const provider = indexById(fixtures.governance.cases).get(
    evidence.providerFixtureId,
  );
  if (
    provider?.providerType !== evidence.providerKind ||
    provider.requiresFile ||
    evidence.sidecarFiles.length > 0
  ) {
    fail("evidence:non-file-provider-invalid");
  }

  const publishedReferences = evidence.publishedRecords.map(
    ({reference}) => reference,
  );
  if (
    new Set(publishedReferences).size !== publishedReferences.length ||
    evidence.publishedRecords.some(
      (record) =>
        !isOpaqueReference(record.reference) ||
        record.subject !== scenario.subject.workspace ||
        record.revision !== scenario.subject.revision,
    )
  ) {
    fail("evidence:published-record-invalid");
  }

  const recovery = evidence.freshContextRecovery;
  if (
    !recovery.newProcess ||
    recovery.originalConversationAvailable ||
    recovery.localSidecarAvailable ||
    recovery.outcome !== "reconstructed" ||
    recovery.resolvedSubject !== scenario.subject.workspace ||
    recovery.resolvedRevision !== scenario.subject.revision ||
    recovery.inputReferences.length === 0 ||
    !recovery.inputReferences.every((reference) =>
      publishedReferences.includes(reference),
    )
  ) {
    fail("evidence:fresh-context-recovery-failed");
  }
};

const validateExternalAndEnterprise = (fixtures, scenario, fail) => {
  const external = scenario.externalEnforcement;
  const composition = indexById(fixtures.external.compositionCases).get(
    external.compositionFixtureId,
  );
  const platform = indexById(fixtures.external.platformFixtures).get(
    external.platformFixtureId,
  );
  if (
    !composition?.expectedValid ||
    !platform?.expectedValid ||
    external.harnessHookEnabled ||
    !external.exactRevision ||
    external.requiredCheck !== platform.requiredCheck ||
    external.checkConclusion !== "failure" ||
    !isOpaqueReference(external.checkReference) ||
    external.targetMutation !== "denied" ||
    external.failureBehavior !== "fail-closed"
  ) {
    fail("external:independent-enforcement-failed");
  }

  const enterprise = scenario.enterpriseComposition;
  const fixturePlatforms = Object.values(
    fixtures.enterprise.mixedStack.providers,
  ).map(({platform: platformId}) => platformId);
  const recordedPlatforms = enterprise.providerReferences.map(
    ({platform: platformId}) => platformId,
  );
  if (
    enterprise.centralResultStore ||
    !enterprise.freshContextRecovery ||
    !hasSameMembers(recordedPlatforms, fixturePlatforms) ||
    enterprise.providerReferences.some(
      ({reference}) => !isOpaqueReference(reference),
    ) ||
    enterprise.federationCaseId !==
      scenario.contractReferences.enterpriseComposition.federationCaseId ||
    enterprise.telemetryCaseId !==
      scenario.contractReferences.enterpriseComposition.telemetryCaseId
  ) {
    fail("enterprise:mixed-stack-recovery-failed");
  }
};

const validateDriftAndRollback = (scenario, fail) => {
  const drift = scenario.drift;
  if (
    !drift.detected ||
    !drift.reported ||
    drift.intendedDigest === drift.observedDigest ||
    drift.classification !== "weakening" ||
    !isOpaqueReference(drift.resultReference) ||
    !["applied", "waived"].includes(drift.remediation.outcome) ||
    !drift.remediation.recommendation ||
    !isOpaqueReference(drift.remediation.previewReference) ||
    !isOpaqueReference(drift.remediation.authorizationReference) ||
    !isOpaqueReference(drift.remediation.resultReference)
  ) {
    fail("drift:remediation-result-invalid");
  }

  const rollback = scenario.rollback;
  if (
    !isOpaqueReference(rollback.authorizationReference) ||
    rollback.targetReference !== scenario.preview.rollback.targetReference ||
    !hasSameMembers(
      rollback.removedModules,
      scenario.application.installedModules,
    ) ||
    !rollback.nativeReferences.every(isOpaqueReference) ||
    !rollback.verified ||
    rollback.installedModulesAfterRollback.length > 0 ||
    rollback.driftDetectedAfterRollback ||
    !rollback.evidenceRetained
  ) {
    fail("rollback:verification-failed");
  }
};

const expectedNegativeResult = (testCase, matrices) => {
  if (
    testCase.kind === "module-trust" &&
    testCase.executable &&
    testCase.authorityZone === "project" &&
    (!testCase.projectTrusted || !testCase.reviewed)
  ) {
    return {outcome: "deny", reasonCode: "project-trust-required"};
  }
  if (testCase.kind === "capability" && testCase.mandatory) {
    const capability = findCapability(
      matrices,
      testCase.platform,
      testCase.capability,
    );
    if (capability?.support !== "supported") {
      return {outcome: "deny", reasonCode: "unsupported-capability"};
    }
  }
  if (
    testCase.kind === "policy-outage" &&
    testCase.mandatory &&
    !testCase.authenticatedCacheAvailable
  ) {
    return {outcome: "deny", reasonCode: "policy-service-unavailable"};
  }
  if (
    testCase.kind === "duplicate-event" &&
    testCase.invocations > 1 &&
    testCase.idempotencyKey &&
    testCase.sideEffectCount === 1
  ) {
    return {outcome: "reconciled", reasonCode: "duplicate-reconciled"};
  }
  if (
    testCase.kind === "agent-recursion" &&
    testCase.requestedDepth > testCase.maximumDepth
  ) {
    return {outcome: "deny", reasonCode: "recursion-denied"};
  }
  if (
    testCase.kind === "exception-expiry" &&
    Date.parse(testCase.expiresAt) < Date.parse(testCase.evaluationTime)
  ) {
    return {outcome: "deny", reasonCode: "exception-expired"};
  }
  return null;
};

const validateNegativeCases = (fixtures, negativeCases, fail) => {
  if (
    !hasSameMembers(
      negativeCases.map(({kind}) => kind),
      requiredNegativeKinds,
    )
  ) {
    fail("negative:required-case-missing");
  }
  const matrices = indexMatrices(fixtures.harness);
  for (const testCase of negativeCases) {
    const expected = expectedNegativeResult(testCase, matrices);
    if (
      !expected ||
      testCase.observedOutcome !== expected.outcome ||
      testCase.reasonCode !== expected.reasonCode
    ) {
      fail(`negative:unsafe-outcome:${testCase.id}`);
    }
  }
};

export const validateWalkingSkeleton = (fixtures) => {
  const failures = [];
  const fail = (message) => failures.push(message);
  const {scenario, negativeCases} = fixtures.walking;

  validateContractReferences(fixtures, scenario, fail);
  validateDiscovery(fixtures, scenario, fail);
  validateAdoptionPaths(fixtures, scenario, fail);
  validateRecommendation(fixtures, scenario, fail);
  validatePreviewAndApplication(fixtures, scenario, fail);
  validateExecution(fixtures, scenario, fail);
  validateEvidence(fixtures, scenario, fail);
  validateExternalAndEnterprise(fixtures, scenario, fail);
  validateDriftAndRollback(scenario, fail);
  validateNegativeCases(fixtures, negativeCases, fail);

  return failures;
};
