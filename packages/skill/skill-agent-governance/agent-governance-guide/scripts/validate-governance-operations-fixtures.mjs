import {readFileSync} from "node:fs";
import {fileURLToPath} from "node:url";
import {
  validateCatalogEntry,
  validateCrossPlatformCase,
  validateDataGovernance,
  validateDriftCase,
  validateLearningCandidate,
  validateMetric,
  validateOperationalOwnership,
  validateOperationalVocabulary,
  validatePolicyBundle,
  validateUpdateCase,
} from "./governance-operations-helpers.mjs";

const fixtureUrl = new URL(
  "../assets/governance-operations-fixtures.json",
  import.meta.url,
);
const fixture = JSON.parse(readFileSync(fileURLToPath(fixtureUrl), "utf8"));

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

validateOperationalVocabulary(fixture);

const failures = [
  ...evaluate(fixture.policyBundles, validatePolicyBundle),
  ...evaluate(fixture.dataCases, validateDataGovernance),
  ...evaluate(fixture.catalogCases, (testCase) =>
    validateCatalogEntry(testCase.entry, testCase.selectedIds),
  ),
  ...evaluate(fixture.driftCases, validateDriftCase),
  ...evaluate(fixture.updateCases, validateUpdateCase),
  ...evaluate(fixture.learningCases, validateLearningCandidate),
  ...evaluate(fixture.metricCases, validateMetric),
  ...evaluate(fixture.operationalCases, validateOperationalOwnership),
  ...evaluate(fixture.crossPlatformCases, validateCrossPlatformCase),
];

if (failures.length > 0) {
  throw new Error(
    `governance operations fixture failures:\n${failures.join("\n")}`,
  );
}

console.log(
  `governance operations fixtures valid: ${fixture.policyBundles.length} policy bundles, ${fixture.dataCases.length} data cases, ${fixture.catalogCases.length} catalog cases, ${fixture.driftCases.length} drift cases, ${fixture.updateCases.length} update cases, ${fixture.learningCases.length} learning cases, ${fixture.metricCases.length} metric cases, ${fixture.operationalCases.length} operational cases, ${fixture.crossPlatformCases.length} cross-platform cases`,
);
