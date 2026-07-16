import {readFileSync} from "node:fs";
import {fileURLToPath} from "node:url";

const fixtureUrl = new URL(
  "../assets/enterprise-platform-fixtures.json",
  import.meta.url,
);
const fixture = JSON.parse(readFileSync(fileURLToPath(fixtureUrl), "utf8"));

const requiredPlatforms = new Set([
  "azure-devops",
  "bitbucket",
  "bitrise",
  "aws",
  "datadog",
]);
const requiredSteps = [
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
const requiredPreviewFields = [
  "nativeChanges",
  "permissions",
  "credentialFlow",
  "networkDestinations",
  "dataCategories",
  "costImpact",
  "failureBehavior",
  "evidence",
  "verification",
  "rollback",
  "driftOwner",
];

const failures = [];
const fail = (message) => failures.push(message);

const platformIds = new Set(fixture.platforms.map(({id}) => id));
for (const id of requiredPlatforms) {
  if (!platformIds.has(id)) fail(`missing platform: ${id}`);
}

for (const platform of fixture.platforms) {
  if (!platform.owner?.endsWith("-guide")) {
    fail(`${platform.id}: missing independent skill owner`);
  }
  if (!platform.sourceRefs?.length) {
    fail(`${platform.id}: missing official source references`);
  }
  for (const variant of platform.variants ?? []) {
    if (
      !variant.product ||
      !variant.testedVersion ||
      !variant.apiVersion ||
      !variant.supportState ||
      !variant.capabilities?.length ||
      !variant.nativeReferences?.length
    ) {
      fail(
        `${platform.id}/${variant.product ?? "unknown"}: incomplete baseline`,
      );
    }
  }
}

for (const step of requiredSteps) {
  if (!fixture.onboarding.steps.includes(step)) {
    fail(`onboarding: missing ${step}`);
  }
}
for (const field of requiredPreviewFields) {
  if (!fixture.onboarding.previewFields.includes(field)) {
    fail(`onboarding preview: missing ${field}`);
  }
}
for (const property of [
  "authorizationBoundToPreview",
  "applyIdempotent",
  "verifyAuthoritativeState",
  "rollbackRetainsEvidence",
  "driftDetection",
]) {
  if (!fixture.onboarding[property])
    fail(`onboarding: ${property} must be true`);
}

const validateFederation = (testCase) => {
  if (testCase.createsLongLivedAccessKey || !testCase.temporaryCredentials) {
    return "long-lived-key-default";
  }
  if (
    !testCase.issuerPinned ||
    !testCase.audienceConstrained ||
    !testCase.subjectConstrained
  ) {
    return "unconstrained-oidc-claims";
  }
  if (
    !testCase.leastPrivilegeRole ||
    !testCase.preview ||
    !testCase.verifyExpiry ||
    !testCase.rollback
  ) {
    return "incomplete-federation-onboarding";
  }
  return null;
};

const validateTelemetry = (testCase) => {
  if (
    !testCase.purposeDeclared ||
    testCase.contentCaptureDefault ||
    !testCase.redaction ||
    !testCase.sampling ||
    !testCase.retention ||
    !testCase.residency ||
    !testCase.access ||
    !testCase.deletion ||
    !testCase.costBudget
  ) {
    return "unsafe-telemetry-default";
  }
  return null;
};

const evaluate = (items, validate) => {
  for (const item of items) {
    const code = validate(item);
    const valid = code === null;
    if (
      valid !== item.expectedValid ||
      (!item.expectedValid && code !== item.expectedCode)
    ) {
      fail(
        `${item.id}: expected valid=${item.expectedValid} code=${item.expectedCode ?? "none"}, received valid=${valid} code=${code ?? "none"}`,
      );
    }
  }
};

evaluate(fixture.federationCases, validateFederation);
evaluate(fixture.telemetryCases, validateTelemetry);

for (const testCase of fixture.negativeCases) {
  if (testCase.silentFallback || !testCase.expectedOutcome) {
    fail(`${testCase.id}: negative case permits silent fallback`);
  }
}

const mixedProviders = Object.values(fixture.mixedStack.providers);
if (
  fixture.mixedStack.centralResultStore ||
  !fixture.mixedStack.freshContextRecovery ||
  fixture.mixedStack.traceIsWorkflowIdentity ||
  mixedProviders.length !== requiredPlatforms.size ||
  mixedProviders.some(({opaque}) => !opaque)
) {
  fail(
    "mixed stack must recover opaque native references without a central store",
  );
}
if (
  new Set(mixedProviders.map(({platform}) => platform)).size !==
  requiredPlatforms.size
) {
  fail("mixed stack must retain five independent provider owners");
}

if (failures.length > 0) {
  throw new Error(
    `enterprise platform fixture failures:\n${failures.join("\n")}`,
  );
}

console.log(
  `enterprise platform fixtures valid: ${fixture.platforms.length} platform owners, ${fixture.platforms.flatMap(({variants}) => variants).length} pinned variants, ${fixture.federationCases.length} federation cases, ${fixture.telemetryCases.length} telemetry cases, ${fixture.negativeCases.length} negative cases, 1 mixed-stack composition`,
);
