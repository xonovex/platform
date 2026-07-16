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
} from "./development-assurance-helpers.mjs";

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
    return validateAssurance(testCase.assurance);
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

export const validateDevelopmentAssuranceFixtures = () => {
  const fixture = JSON.parse(readFileSync(fileURLToPath(fixtureUrl), "utf8"));
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

  if (failures.length > 0) {
    throw new Error(
      `development and assurance fixture failures:\n${failures.join("\n")}`,
    );
  }

  return {
    cases: fixture.cases.length,
    inventorySpecializations: fixture.inventorySpecializations.length,
  };
};
