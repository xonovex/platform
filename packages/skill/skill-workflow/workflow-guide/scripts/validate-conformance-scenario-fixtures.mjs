import {existsSync, readdirSync, readFileSync} from "node:fs";
import {fileURLToPath} from "node:url";

const fixturesUrl = new URL("../assets/fixtures/", import.meta.url);
const repoRootUrl = new URL("../../../../../", import.meta.url);

const isNonEmptyString = (value) =>
  typeof value === "string" && value.trim().length > 0;

const isNonEmptyStringArray = (value) =>
  Array.isArray(value) && value.length > 0 && value.every(isNonEmptyString);

const validateScenario = (fileName, fixture, entry) => {
  const failures = [];
  const stem = fileName.replace(/\.json$/, "");
  if (fixture.id !== stem) {
    failures.push(`${fileName}: id '${fixture.id}' does not match file name`);
  }
  if (!isNonEmptyString(fixture.contract)) {
    failures.push(`${fileName}: contract is missing or empty`);
  }
  if (!isNonEmptyString(fixture.scenario)) {
    failures.push(`${fileName}: scenario is missing or empty`);
  }
  if (
    fixture.given === undefined ||
    fixture.given === null ||
    typeof fixture.given !== "object"
  ) {
    failures.push(`${fileName}: given must be an object or array`);
  }
  if (!isNonEmptyString(fixture.when)) {
    failures.push(`${fileName}: when is missing or empty`);
  }
  if (typeof fixture.expect !== "object" || fixture.expect === null) {
    failures.push(`${fileName}: expect is missing`);
    return failures;
  }
  if (!isNonEmptyString(fixture.expect.outcome)) {
    failures.push(`${fileName}: expect.outcome is missing or empty`);
  }
  if (!isNonEmptyStringArray(fixture.expect.reasons)) {
    failures.push(
      `${fileName}: expect.reasons must be a non-empty string array`,
    );
  }
  if (!isNonEmptyStringArray(fixture.expect.must_not)) {
    failures.push(
      `${fileName}: expect.must_not must be a non-empty string array`,
    );
  }
  if (entry === undefined) {
    failures.push(`${fileName}: fixture has no index.json entry`);
    return failures;
  }
  const referenceStem = entry.reference.split("/").at(-1)?.replace(/\.md$/, "");
  if (fixture.contract !== referenceStem) {
    failures.push(
      `${fileName}: contract '${fixture.contract}' does not match owning reference '${entry.reference}'`,
    );
  }
  if (!existsSync(new URL(entry.reference, repoRootUrl))) {
    failures.push(
      `${fileName}: owning reference '${entry.reference}' does not exist`,
    );
  }
  return failures;
};

const validateCorpus = (fileNames, readFixture, index) => {
  const failures = [];
  const indexed = new Map(index.fixtures.map((entry) => [entry.file, entry]));
  for (const entry of index.fixtures) {
    if (!fileNames.includes(entry.file)) {
      failures.push(`index.json: entry '${entry.file}' has no fixture file`);
    }
  }
  if (indexed.size !== index.fixtures.length) {
    failures.push("index.json: duplicate fixture entries");
  }
  for (const fileName of fileNames) {
    failures.push(
      ...validateScenario(
        fileName,
        readFixture(fileName),
        indexed.get(fileName),
      ),
    );
  }
  return failures;
};

const fixtureFiles = readdirSync(fileURLToPath(fixturesUrl))
  .filter((name) => name.endsWith(".json") && name !== "index.json")
  .sort();
const readFixture = (name) =>
  JSON.parse(readFileSync(new URL(name, fixturesUrl), "utf8"));
const index = JSON.parse(
  readFileSync(new URL("index.json", fixturesUrl), "utf8"),
);

const failures = validateCorpus(fixtureFiles, readFixture, index);
if (failures.length > 0) {
  throw new Error(`conformance scenario failures:\n${failures.join("\n")}`);
}

const guards = [
  validateCorpus(
    fixtureFiles,
    (name) =>
      name === fixtureFiles[0]
        ? {...readFixture(name), expect: {outcome: "report"}}
        : readFixture(name),
    index,
  ),
  validateCorpus(fixtureFiles, readFixture, {
    fixtures: index.fixtures.slice(1),
  }),
  validateCorpus(
    fixtureFiles,
    (name) =>
      name === fixtureFiles[0]
        ? {...readFixture(name), contract: "not-the-owner"}
        : readFixture(name),
    index,
  ),
  validateCorpus(fixtureFiles, readFixture, {
    fixtures: [
      ...index.fixtures,
      {file: "ghost.json", reference: index.fixtures[0].reference},
    ],
  }),
];
const dudGuards = guards.filter((guardFailures) => guardFailures.length === 0);
if (dudGuards.length > 0) {
  throw new Error(
    `conformance scenario mutation guards failed: ${dudGuards.length} tampered corpora passed validation`,
  );
}

console.log(
  `conformance scenarios valid: ${fixtureFiles.length} fixtures, ${new Set(index.fixtures.map((entry) => entry.reference)).size} owning references, ${guards.length} mutation guards`,
);
