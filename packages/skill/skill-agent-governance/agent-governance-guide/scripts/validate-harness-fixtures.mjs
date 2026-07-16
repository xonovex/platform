import {readFileSync} from "node:fs";
import {fileURLToPath} from "node:url";
import {
  indexMatrices,
  validateHarnessCase,
  validateMatrix,
  validateTemplates,
} from "./harness-conformance-helpers.mjs";

const readJson = (relativePath) =>
  JSON.parse(
    readFileSync(fileURLToPath(new URL(relativePath, import.meta.url)), "utf8"),
  );

const fixture = readJson("../assets/harness-conformance-fixtures.json");
const templates = readJson("../assets/harness-module-templates.json");
const matrices = indexMatrices(fixture);

const matrixFailures = fixture.matrices.flatMap((matrix) => {
  const code = validateMatrix(matrix, fixture);
  return code ? [`${matrix.platform}:${code}`] : [];
});
const duplicatePlatforms =
  matrices.size === fixture.matrices.length
    ? []
    : ["duplicate-platform-matrix"];

const caseFailures = fixture.cases.flatMap((testCase) => {
  const code = validateHarnessCase(testCase, matrices);
  const valid = code === null;
  const codeMatches = testCase.expectedValid || code === testCase.expectedCode;
  return valid === testCase.expectedValid && codeMatches
    ? []
    : [
        `${testCase.id}: expected valid=${testCase.expectedValid} code=${testCase.expectedCode ?? "none"}, received valid=${valid} code=${code ?? "none"}`,
      ];
});

const templateFailures = validateTemplates(
  templates,
  fixture.matrices.map(({platform}) => platform),
);
const failures = [
  ...matrixFailures,
  ...duplicatePlatforms,
  ...caseFailures,
  ...templateFailures,
];

if (failures.length > 0) {
  throw new Error(`harness fixture failures:\n${failures.join("\n")}`);
}

console.log(
  `harness fixtures valid: ${fixture.matrices.length} platform matrices, ${fixture.cases.length} semantic cases, ${templates.deterministic.length} deterministic templates, 1 bounded model evaluator, 1 bounded agent launcher`,
);
