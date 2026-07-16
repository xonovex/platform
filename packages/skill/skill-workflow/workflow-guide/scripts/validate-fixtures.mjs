import {readFileSync} from "node:fs";
import {fileURLToPath} from "node:url";
import {
  exerciseTaskSystemProvider,
  validateResultContracts,
  validateResultProvider,
  validateWorkflowCase,
} from "./conformance-helpers.mjs";
import {validateDevelopmentAssuranceFixtures} from "./validate-development-assurance-fixtures.mjs";

const fixtureUrl = new URL(
  "../assets/conformance-fixtures.json",
  import.meta.url,
);
const fixture = JSON.parse(readFileSync(fileURLToPath(fixtureUrl), "utf8"));

const expectedKinds = [
  "Discovery",
  "Research",
  "Formulation",
  "ExperienceDesign",
  "SolutionDesign",
  "Decision",
  "Planning",
  "Development",
  "DeliverablePublication",
  "Review",
  "QA",
  "Assessment",
  "Inventory",
  "DataGovernance",
  "Acceptance",
  "Integration",
  "Transition",
  "Release",
  "Observation",
  "Incident",
  "CorrectiveAction",
  "Retirement",
  "Learning",
];

validateResultContracts(
  fixture.resultContracts,
  expectedKinds,
  fixture.inventoryOptionalComponents,
);

const providerFailures = fixture.providerFixtures.flatMap((provider) => {
  const code = validateResultProvider(provider);
  const valid = code === null;
  const expectedCodeMatches =
    provider.expectedValid || code === provider.expectedCode;
  return valid === provider.expectedValid && expectedCodeMatches
    ? []
    : [
        `${provider.id}: expected valid=${provider.expectedValid} code=${provider.expectedCode ?? "none"}, received valid=${valid} code=${code ?? "none"}`,
      ];
});

exerciseTaskSystemProvider();

const developmentAssurance = validateDevelopmentAssuranceFixtures();

const failures = fixture.cases.flatMap((testCase) => {
  const code = validateWorkflowCase(testCase);
  const valid = code === null;
  const expectedCodeMatches =
    testCase.expectedValid || code === testCase.expectedCode;
  return valid === testCase.expectedValid && expectedCodeMatches
    ? []
    : [
        `${testCase.id}: expected valid=${testCase.expectedValid} code=${testCase.expectedCode ?? "none"}, received valid=${valid} code=${code ?? "none"}`,
      ];
});

if (providerFailures.length > 0 || failures.length > 0) {
  throw new Error(
    `workflow fixture failures:\n${[...providerFailures, ...failures].join("\n")}`,
  );
}

console.log(
  `workflow fixtures valid: ${fixture.resultContracts.length} result contracts, ${fixture.providerFixtures.length} providers, ${fixture.cases.length} cases`,
);
console.log(
  `development and assurance fixtures valid: ${developmentAssurance.cases} cases, ${developmentAssurance.inventorySpecializations} inventory specializations`,
);
