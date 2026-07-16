import {readFileSync} from "node:fs";
import {fileURLToPath} from "node:url";
import {
  mergeFixture,
  validateAgent,
  validateBreakGlass,
  validateComposition,
  validateConfigurationCase,
  validateEnforcement,
  validateEvidenceCase,
  validateException,
  validateExecutor,
  validateModule,
  validateNativeProvider,
  validateOrdering,
  validatePolicy,
  validatePolicyParity,
  validateProfile,
  validateProviderContract,
  validateVocabulary,
} from "./conformance-helpers.mjs";

const fixtureUrl = new URL(
  "../assets/conformance-fixtures.json",
  import.meta.url,
);
const fixture = JSON.parse(readFileSync(fileURLToPath(fixtureUrl), "utf8"));
const evaluationTime = Date.parse(fixture.evaluationTime);

const validators = {
  executor: validateExecutor,
  enforcement: validateEnforcement,
  composition: validateComposition,
  ordering: validateOrdering,
  agent: validateAgent,
  policy: validatePolicy,
  exception: (testCase) => validateException(testCase, evaluationTime),
  "break-glass": (testCase) => validateBreakGlass(testCase, evaluationTime),
  provider: validateNativeProvider,
  profile: validateProfile,
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

validateVocabulary(fixture);

const providerFailures = evaluate(
  fixture.providerContracts,
  validateProviderContract,
);
const policyFailures = evaluate(fixture.policyCases, validatePolicyParity);
const evidenceFailures = evaluate(fixture.evidenceCases, validateEvidenceCase);
const configurationFailures = evaluate(
  fixture.configurationCases,
  validateConfigurationCase,
);
const moduleFailures = evaluate(fixture.moduleCases, (testCase) =>
  validateModule(
    mergeFixture(
      fixture.moduleFixtures[testCase.fixture],
      testCase.overrides ?? {},
    ),
    fixture.moduleRequiredFields,
  ),
);
const contractFailures = evaluate(fixture.cases, (testCase) => {
  const validator = validators[testCase.kind];
  return validator ? validator(testCase) : "unknown-contract-kind";
});

const failures = [
  ...providerFailures,
  ...policyFailures,
  ...evidenceFailures,
  ...configurationFailures,
  ...moduleFailures,
  ...contractFailures,
];

if (failures.length > 0) {
  throw new Error(`governance fixture failures:\n${failures.join("\n")}`);
}

console.log(
  `governance fixtures valid: ${fixture.providerContracts.length} provider contracts, ${fixture.policyCases.length} policy parity cases, ${fixture.configurationCases.length} configuration cases, ${fixture.moduleCases.length} module cases, ${fixture.cases.length} composition cases`,
);
