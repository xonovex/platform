import {readFileSync} from "node:fs";
import {fileURLToPath} from "node:url";

const fixtureUrl = new URL(
  "../assets/external-enforcement-fixtures.json",
  import.meta.url,
);
const fixture = JSON.parse(readFileSync(fileURLToPath(fixtureUrl), "utf8"));

const fullCommitSha = /^[0-9a-f]{40}$/u;
const externalPoints = new Set([
  "required-ci-check",
  "protected-environment",
  "provider-permission",
  "repository-ruleset",
  "pipeline-execution-policy",
  "admission-webhook",
]);

const validatePlatform = (testCase) => {
  if (testCase.platform === "github") {
    if (!fullCommitSha.test(testCase.moduleRef)) {
      return "immutable-module-pin";
    }
    if (!testCase.requiredCheck || !testCase.rulesetActive) {
      return "required-native-control";
    }
    if (!testCase.environment?.requiredReviewer) {
      return "protected-environment-approval";
    }
    if (!testCase.evidence?.includes("check-run")) {
      return "native-evidence-required";
    }
    return null;
  }

  if (testCase.platform === "gitlab") {
    if (!fullCommitSha.test(testCase.componentRef)) {
      return "immutable-module-pin";
    }
    if (!testCase.typedInputs || !testCase.componentTested) {
      return "component-contract";
    }
    if (!testCase.configurableJobName) {
      return "ambiguous-job-name";
    }
    if (!testCase.pipelineExecutionPolicy || !testCase.protectedEnvironment) {
      return "required-native-control";
    }
    return null;
  }

  if (testCase.platform === "kubernetes") {
    if (
      testCase.agentPolicyCount !== 1 ||
      testCase.policyLookupFailure !== "deny"
    ) {
      return "namespace-policy-authority";
    }
    if (
      !testCase.runtimeClass ||
      !testCase.securityContextRequired ||
      !testCase.networkPolicyRequired ||
      !testCase.maxTimeout ||
      !testCase.maxResources
    ) {
      return "admission-intent-coverage";
    }
    if (testCase.attempt?.expectedOutcome !== "deny") {
      return "bypass-attempt-not-denied";
    }
    if (!testCase.toolchainPinned || !testCase.namespaceQuota) {
      return "independent-layer-required";
    }
    return null;
  }

  return "unknown-platform";
};

const validateComposition = (testCase) => {
  const points = new Set(testCase.enforcementPoints);
  const hasExternalPoint = [...points].some((point) =>
    externalPoints.has(point),
  );

  if (testCase.mandatory && points.has("client-hook")) {
    return "bypassable-enforcement";
  }
  if (testCase.mandatory && points.has("harness-hook") && !hasExternalPoint) {
    return "independent-layer-required";
  }
  if (testCase.mandatory && !testCase.exactRevision) {
    return "exact-revision-required";
  }
  if (testCase.mandatory && testCase.failureBehavior !== "fail-closed") {
    return "mandatory-failure-policy";
  }
  if (
    testCase.privilegedOperation &&
    (!hasExternalPoint || !testCase.authorizedActor || !testCase.rollback)
  ) {
    return "privileged-gate-required";
  }
  return null;
};

const validatePolicyService = (testCase) => {
  if (testCase.mandatory && testCase.outageOutcome !== "deny") {
    return "mandatory-outage-must-deny";
  }
  if (
    !testCase.cache?.authenticated ||
    !testCase.cache?.subjectBound ||
    !testCase.cache?.policyVersion ||
    testCase.cache?.expired
  ) {
    return "invalid-decision-cache";
  }
  if (
    !testCase.historicalEvidence?.decisionRef ||
    !testCase.historicalEvidence?.inputDigest ||
    testCase.historicalEvidence?.policyVersion !== testCase.cache.policyVersion
  ) {
    return "historical-policy-evidence";
  }
  return null;
};

const onboardingSteps = [
  "discover",
  "assess",
  "propose",
  "preview",
  "authorize",
  "apply",
  "verify",
  "record",
  "operate",
];

const validateOnboarding = (testCase) => {
  if (
    onboardingSteps.some((step) => !testCase.steps.includes(step)) ||
    !testCase.nativeChanges ||
    !testCase.permissionsShown ||
    !testCase.dataFlowsShown ||
    !testCase.rollback ||
    !testCase.negativeProbe ||
    !testCase.driftOwner
  ) {
    return "incomplete-onboarding-transaction";
  }
  return null;
};

const evaluate = (items, validate) =>
  items.flatMap((item) => {
    const code = validate(item);
    const valid = code === null;
    const expectedCodeMatches =
      item.expectedValid || code === item.expectedCode;
    return valid === item.expectedValid && expectedCodeMatches
      ? []
      : [
          `${item.id}: expected valid=${item.expectedValid} code=${item.expectedCode ?? "none"}, received valid=${valid} code=${code ?? "none"}`,
        ];
  });

const failures = [
  ...evaluate(fixture.platformFixtures, validatePlatform),
  ...evaluate(fixture.compositionCases, validateComposition),
  ...evaluate(fixture.policyServiceCases, validatePolicyService),
  ...evaluate(fixture.onboardingCases, validateOnboarding),
];

if (failures.length > 0) {
  throw new Error(
    `external enforcement fixture failures:\n${failures.join("\n")}`,
  );
}

console.log(
  `external enforcement fixtures valid: ${fixture.platformFixtures.length} platform fixtures, ${fixture.compositionCases.length} composition cases, ${fixture.policyServiceCases.length} policy-service cases, ${fixture.onboardingCases.length} onboarding cases`,
);
