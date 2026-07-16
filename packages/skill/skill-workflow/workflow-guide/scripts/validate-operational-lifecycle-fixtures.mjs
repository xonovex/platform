import {readFileSync} from "node:fs";
import {fileURLToPath} from "node:url";
import {
  validateAcceptance,
  validateAgentAssistance,
  validateAuthorization,
  validateCorrectiveAction,
  validateEmergencyAccess,
  validateIncident,
  validateObservation,
  validatePrivilegedOperation,
  validateRelease,
  validateRetirement,
  validateTransition,
} from "./operational-lifecycle-helpers.mjs";

const fixtureUrl = new URL(
  "../assets/operational-lifecycle-fixtures.json",
  import.meta.url,
);

const isRecord = (value) =>
  value !== null && typeof value === "object" && !Array.isArray(value);

const mergeFixture = (base, override) => {
  if (!isRecord(base) || !isRecord(override)) {
    return override === undefined ? base : override;
  }
  return Object.fromEntries(
    [...new Set([...Object.keys(base), ...Object.keys(override)])].map(
      (key) => [key, mergeFixture(base[key], override[key])],
    ),
  );
};

const validateCase = (testCase, templates) => {
  const input = mergeFixture(
    templates[testCase.template] ?? {},
    testCase.input,
  );
  if (testCase.contract === "acceptance") {
    return validateAcceptance(input);
  }
  if (testCase.contract === "authorization") {
    return validateAuthorization(input);
  }
  if (testCase.contract === "emergency-access") {
    return validateEmergencyAccess(input);
  }
  if (testCase.contract === "privileged-operation") {
    return validatePrivilegedOperation(input);
  }
  if (testCase.contract === "transition") {
    return validateTransition(input);
  }
  if (testCase.contract === "release") {
    return validateRelease(input);
  }
  if (testCase.contract === "observation") {
    return validateObservation(input);
  }
  if (testCase.contract === "incident") {
    return validateIncident(input);
  }
  if (testCase.contract === "corrective-action") {
    return validateCorrectiveAction(input);
  }
  if (testCase.contract === "retirement") {
    return validateRetirement(input);
  }
  if (testCase.contract === "agent-assistance") {
    return validateAgentAssistance(input);
  }
  return `unknown-contract:${testCase.contract}`;
};

export const validateOperationalLifecycleFixtures = () => {
  const fixture = JSON.parse(readFileSync(fileURLToPath(fixtureUrl), "utf8"));
  const failures = fixture.cases.flatMap((testCase) => {
    const code = validateCase(testCase, fixture.templates);
    const valid = code === null;
    const expectedCodeMatches =
      testCase.expectedValid || code === testCase.expectedCode;
    return valid === testCase.expectedValid && expectedCodeMatches
      ? []
      : [
          `${testCase.id}: expected valid=${testCase.expectedValid} code=${testCase.expectedCode ?? "none"}, received valid=${valid} code=${code ?? "none"}`,
        ];
  });

  if (failures.length > 0) {
    throw new Error(
      `operational lifecycle fixture failures:\n${failures.join("\n")}`,
    );
  }

  return {cases: fixture.cases.length};
};
