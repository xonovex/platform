import {readFileSync} from "node:fs";
import {fileURLToPath} from "node:url";
import {
  evaluateFreshness,
  selectDevelopmentExecutor,
  validateAssurance,
  validateAssuranceBatch,
  validateDeliverable,
  validateDevelopment,
  validateExternalEvidence,
  validateInventory,
} from "./development-assurance-helpers.ts";

const fixtureUrl = new URL(
  "../assets/development-assurance-fixtures.json",
  import.meta.url,
);

const validateCase = (testCase, inventorySpecializations) => {
  if (testCase.contract === "executor-selection") {
    const result = selectDevelopmentExecutor(testCase.input);
    return (
      result.code ??
      (result.executor === testCase.expectedExecutor
        ? null
        : `unexpected-executor:${result.executor}`)
    );
  }
  if (testCase.contract === "development") {
    return validateDevelopment(testCase.development);
  }
  if (testCase.contract === "deliverable") {
    return validateDeliverable(testCase.deliverable);
  }
  if (testCase.contract === "inventory") {
    return validateInventory(testCase.inventory, inventorySpecializations);
  }
  if (testCase.contract === "assurance") {
    return validateAssurance({
      assurance: testCase.assurance,
      profile: testCase.profile,
    });
  }
  if (testCase.contract === "assurance-batch") {
    return validateAssuranceBatch(testCase.assuranceBatch);
  }
  if (testCase.contract === "freshness") {
    return evaluateFreshness(testCase.freshness);
  }
  if (testCase.contract === "external-evidence") {
    return validateExternalEvidence(testCase.externalEvidence);
  }
  return `unknown-contract:${testCase.contract}`;
};

// Review independence is resolved from the profile, so these cases prove an
// assessor cannot elect the check applied to them. Deleting one would silently
// reopen the self-assessment path; requiredCaseIds keeps them under test.
//
// `distinct-team` cannot be satisfied by the identity comparison alone, so three
// cases pin the provider-evidence requirement on the Review path: a distinct
// assessor with the evidence reference recorded passes, the same record without
// it fails closed as `independence-provider-evidence-required` rather than
// downgrading to identity-only, and a self-assessment is still caught as
// `assessor-independence-failed` even when the evidence is present.
const requiredCaseIds = [
  "review-rejects-self-assessment",
  "review-record-cannot-waive-profile-independence",
  "review-independence-requirement-undeclared",
  "review-independence-unverifiable-without-assessor",
  "profile-may-waive-review-independence",
  "review-distinct-team-with-provider-evidence-is-valid",
  "review-distinct-team-without-provider-evidence-fails-closed",
  "review-distinct-team-self-grant-caught-with-provider-evidence",
];

export const validateDevelopmentAssuranceFixtures = () => {
  const fixture = JSON.parse(readFileSync(fileURLToPath(fixtureUrl), "utf8"));
  const presentCaseIds = new Set(fixture.cases.map(({id}) => id));
  const missingCaseIds = requiredCaseIds
    .filter((id) => !presentCaseIds.has(id))
    .map((id) => `required fixture case is missing: ${id}`);
  const failures = fixture.cases.flatMap((testCase) => {
    const code = validateCase(testCase, fixture.inventorySpecializations);
    const valid = code === null;
    const expectedCodeMatches =
      testCase.expectedValid || code === testCase.expectedCode;
    return valid === testCase.expectedValid && expectedCodeMatches
      ? []
      : [
          `${testCase.id}: expected valid=${testCase.expectedValid} code=${testCase.expectedCode ?? "none"}, received valid=${valid} code=${code ?? "none"}`,
        ];
  });

  if (missingCaseIds.length > 0 || failures.length > 0) {
    throw new Error(
      `development and assurance fixture failures:\n${[...missingCaseIds, ...failures].join("\n")}`,
    );
  }

  return {
    cases: fixture.cases.length,
    inventorySpecializations: fixture.inventorySpecializations.length,
  };
};
