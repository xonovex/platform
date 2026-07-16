const supportStates = new Set([
  "supported",
  "unsupported",
  "experimental",
  "unknown",
]);
const coverageStates = new Set(["complete", "partial", "none"]);

const missingFields = (value, fields) =>
  fields.filter(
    (field) =>
      !Object.hasOwn(value, field) ||
      value[field] === null ||
      value[field] === undefined,
  );

export const indexMatrices = (fixture) =>
  new Map(fixture.matrices.map((matrix) => [matrix.platform, matrix]));

export const findCapability = (matrices, platform, capabilityId) =>
  matrices.get(platform)?.capabilities.find(({id}) => id === capabilityId) ??
  null;

export const validateMatrix = (matrix, fixture) => {
  const missing = missingFields(matrix, fixture.matrixRequiredFields);
  if (missing.length > 0) return `matrix-field-missing:${missing.join(",")}`;
  if (!/^\d+\.\d+\.\d+$/.test(matrix.matrixVersion)) {
    return "matrix-version-invalid";
  }
  if (!Number.isFinite(Date.parse(matrix.documentationSnapshot))) {
    return "matrix-snapshot-invalid";
  }
  if (
    !matrix.runtimeProbe.command ||
    !["probed", "not-installed"].includes(matrix.runtimeProbe.status)
  ) {
    return "runtime-probe-invalid";
  }
  if (
    !Array.isArray(matrix.configuration.nativePaths) ||
    matrix.configuration.nativePaths.length === 0 ||
    !Array.isArray(matrix.configuration.scopes) ||
    matrix.configuration.scopes.length === 0
  ) {
    return "native-configuration-incomplete";
  }
  if (!Array.isArray(matrix.capabilities) || matrix.capabilities.length === 0) {
    return "capability-matrix-empty";
  }
  for (const capability of matrix.capabilities) {
    const capabilityMissing = missingFields(
      capability,
      fixture.capabilityRequiredFields,
    );
    if (capabilityMissing.length > 0) {
      return `capability-field-missing:${capability.id ?? "unknown"}:${capabilityMissing.join(",")}`;
    }
    if (
      !supportStates.has(capability.support) ||
      !coverageStates.has(capability.coverage)
    ) {
      return `capability-state-invalid:${capability.id}`;
    }
    if (capability.support === "supported" && !capability.executes) {
      return `supported-handler-does-not-execute:${capability.id}`;
    }
  }
  return null;
};

const validateCapabilityCase = (testCase, matrices) => {
  const capability = findCapability(
    matrices,
    testCase.platform,
    testCase.capability,
  );
  if (
    !capability ||
    capability.support !== "supported" ||
    !capability.executes
  ) {
    return "unsupported-capability";
  }
  if (
    testCase.requiredCoverage === "complete" &&
    capability.coverage !== "complete"
  ) {
    return "partial-coverage";
  }
  if (testCase.requiresBlocking && !capability.blocking) {
    return "nonblocking-capability";
  }
  return null;
};

const validateOrderingCase = (testCase, matrices) => {
  const capability = findCapability(
    matrices,
    testCase.platform,
    testCase.capability,
  );
  if (!capability) return "unknown-capability";
  return testCase.requiresSerial && !capability.ordering.includes("sequential")
    ? "ordering-unguaranteed"
    : null;
};

const validateContextCase = (testCase, matrices) => {
  const capability = findCapability(
    matrices,
    testCase.platform,
    testCase.capability,
  );
  return capability?.support === "supported" && capability.contextInjection
    ? null
    : "context-injection-unsupported";
};

const validatePrecedenceCase = (testCase) =>
  testCase.lowerScopeWeakensMandatory && !testCase.nativePreventsWeakening
    ? "unauthorized-weakening"
    : null;

const validateOnboardingCase = (testCase) => {
  if (
    !testCase.previewed ||
    !testCase.permissionReport ||
    !testCase.dataFlowReport
  ) {
    return "onboarding-preview-incomplete";
  }
  if (testCase.applied && !testCase.authorized) {
    return "onboarding-unauthorized";
  }
  if ((testCase.retryCount ?? 0) > 0 && !testCase.idempotencyKey) {
    return "onboarding-duplicate-risk";
  }
  if (testCase.applied && !testCase.verified) {
    return "onboarding-unverified";
  }
  if (testCase.applied && !testCase.rollbackAvailable) {
    return "onboarding-rollback-unavailable";
  }
  return testCase.driftChecked ? null : "onboarding-drift-unchecked";
};

const validateTrustCase = (testCase) => {
  if (
    testCase.executable &&
    testCase.authorityZone === "project" &&
    (!testCase.projectTrusted || !testCase.reviewed)
  ) {
    return "project-trust-required";
  }
  if (
    testCase.authorityZone === "organization-managed" &&
    (!testCase.sourceRetained || !testCase.provenanceRetained)
  ) {
    return "managed-provenance-incomplete";
  }
  return null;
};

export const validateHarnessCase = (testCase, matrices) => {
  const validators = {
    capability: validateCapabilityCase,
    ordering: validateOrderingCase,
    context: validateContextCase,
    precedence: validatePrecedenceCase,
    onboarding: validateOnboardingCase,
    trust: validateTrustCase,
  };
  const validate = validators[testCase.kind];
  return validate ? validate(testCase, matrices) : "unknown-case-kind";
};

export const validateTemplates = (templates, platforms) => {
  const failures = [];
  for (const template of templates.deterministic) {
    if (
      template.executorClass !== "deterministic" ||
      !Array.isArray(template.inputs) ||
      !Array.isArray(template.outputs) ||
      !Number.isFinite(template.timeoutMs) ||
      template.timeoutMs <= 0 ||
      !["fail-closed", "fail-visible", "advisory"].includes(
        template.failureMode,
      )
    ) {
      failures.push(`deterministic-template-invalid:${template.id}`);
    }
  }

  const evaluator = templates.modelEvaluator;
  if (
    evaluator.executorClass !== "model" ||
    !evaluator.explicitRunner ||
    !evaluator.outputSchema?.required?.length ||
    evaluator.outputSchema.additionalProperties !== false ||
    evaluator.timeBudgetSeconds <= 0 ||
    evaluator.tokenBudget <= 0 ||
    evaluator.costBudget < 0 ||
    evaluator.failureMode !== "fail-visible"
  ) {
    failures.push("model-evaluator-unbounded");
  }

  const agent = templates.agentLauncher;
  if (
    agent.executorClass !== "agent" ||
    !agent.explicitLaunch ||
    agent.maxDepth !== 1 ||
    agent.timeBudgetSeconds <= 0 ||
    agent.tokenBudget <= 0 ||
    agent.costBudget < 0 ||
    !agent.authorityAttenuated ||
    agent.recursion !== "denied" ||
    !agent.killSwitch ||
    agent.failureMode !== "fail-visible"
  ) {
    failures.push("agent-launcher-unbounded");
  }

  const mappingPlatforms = new Set(
    templates.platformMappings.map(({platform}) => platform),
  );
  const missingPlatforms = platforms.filter(
    (platform) => !mappingPlatforms.has(platform),
  );
  if (missingPlatforms.length > 0) {
    failures.push(`template-platform-missing:${missingPlatforms.join(",")}`);
  }
  const nativePaths = new Set(
    templates.platformMappings.map(({nativePath}) => nativePath),
  );
  if (nativePaths.size !== templates.platformMappings.length) {
    failures.push("universal-hook-file-detected");
  }
  return failures;
};
