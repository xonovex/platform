import {readFileSync, statSync} from "node:fs";
import {
  capabilityNames,
  capabilityRegistry,
  registryEntries,
  resolveCapability,
} from "./capability-registry-helpers.mjs";
import {expectedVocabulary} from "./conformance-helpers.mjs";

// Guard the capability registry: every selectable composition capability is
// classified shipped or adopter-supplied (no unclassified entry), the closed
// vocabularies stay consistent with the owners they mirror, every shipped owner
// package exists, the verified placeholder methods stay adopter-supplied, and the
// registry fails closed on an unknown capability so the Phase 3 profile validator
// and the Phase 5 completeness check reject a dangling reference. Mutation guards
// replay each tamper this validator exists to catch; a guard reporting no failure
// (a dud) means the comparison stopped working.

const repoRootUrl = new URL("../../../../../", import.meta.url);

const readSource = (path: string): string =>
  readFileSync(new URL(path, repoRootUrl), "utf8");

const dirExists = (path: string): boolean => {
  try {
    return statSync(new URL(path, repoRootUrl)).isDirectory();
  } catch {
    return false;
  }
};

const absent = (
  members: readonly string[],
  other: readonly string[],
): string[] => members.filter((member) => !other.includes(member));

const validStatuses = ["shipped", "adopter-supplied"];

interface RegistryEntry {
  category: string;
  name: string;
  status: string;
  owner?: string;
  contract?: string;
}

// A single entry is well-classified when its status is known and it names the
// field that status requires — an owner for shipped, a contract for
// adopter-supplied. Returns the failure text, or null when the entry is sound.
const entryFailure = (entry: RegistryEntry): string | null => {
  const where = `${entry.category}/${entry.name}`;
  if (!validStatuses.includes(entry.status)) {
    return `${where}: status "${entry.status}" is neither shipped nor adopter-supplied`;
  }
  if (entry.status === "shipped" && !entry.owner) {
    return `${where}: shipped capability names no owner`;
  }
  if (entry.status === "adopter-supplied" && !entry.contract) {
    return `${where}: adopter-supplied capability names no contract`;
  }
  return null;
};

const compareSet = (
  label: string,
  actual: readonly string[],
  expected: readonly string[],
): string[] => {
  const missing = absent(expected, actual);
  const unexpected = absent(actual, expected);
  const failures: string[] = [];
  if (missing.length > 0)
    failures.push(`${label}: missing ${missing.join(", ")}`);
  if (unexpected.length > 0) {
    failures.push(`${label}: unexpected ${unexpected.join(", ")}`);
  }
  return failures;
};

// The Method axis names its selectable methods in one sentence; normalize each
// to the registry's kebab identifier (lowercase, spaces to hyphens).
const parseMethodAxis = (source: string): string[] | null => {
  const match =
    /neutral is the default\.\s*(.+?),?\s+and other installed skills/.exec(
      source,
    );
  if (match === null) return null;
  const methods = match[1]
    .split(/,\s*/)
    .map((method) => method.trim().toLowerCase().replace(/\s+/g, "-"))
    .filter((method) => method.length > 0);
  return methods.length > 0 ? methods : null;
};

const entries: RegistryEntry[] = registryEntries();
const failures: string[] = [];

// 1. Completeness — every entry classified with the field its status requires.
for (const entry of entries) {
  const failure = entryFailure(entry);
  if (failure !== null) failures.push(failure);
}

// 2. Consistency — the closed vocabularies mirror the owners they derive from.
failures.push(
  ...compareSet(
    "executor-class vs expectedVocabulary.executorClasses",
    capabilityNames("executor-class"),
    expectedVocabulary.executorClasses,
  ),
  ...compareSet(
    "governance-module vs expectedVocabulary.moduleClassifications",
    capabilityNames("governance-module"),
    expectedVocabulary.moduleClassifications,
  ),
);

// 3. Method axis — the registry's methods match the declaring contract, so a
// method added to the doc without classification is caught.
const methodAxis = parseMethodAxis(
  readSource(
    "packages/skill/skill-plan/plan-guide/references/early-lifecycle-contracts.md",
  ),
);
if (methodAxis === null) {
  failures.push(
    "method: could not parse the Method axis list in early-lifecycle-contracts.md",
  );
} else {
  failures.push(
    ...compareSet(
      "method vs early-lifecycle-contracts.md Method axis",
      capabilityNames("method"),
      methodAxis,
    ),
  );
}

// 4. Shipped owners exist, and the verified placeholder methods stay
// adopter-supplied so no bare method implies a shipped skill.
for (const entry of entries) {
  if (entry.status === "shipped" && entry.owner !== undefined) {
    if (!dirExists(`packages/skill/${entry.owner}/`)) {
      failures.push(
        `${entry.category}/${entry.name}: shipped owner ${entry.owner} has no package under packages/skill/`,
      );
    }
  }
}
for (const name of [
  "example-mapping",
  "user-research",
  "architecture-review",
]) {
  const status = resolveCapability("method", name);
  if (status !== "adopter-supplied") {
    failures.push(
      `method/${name}: verified placeholder must be adopter-supplied, got ${status}`,
    );
  }
}

// 5. Fail-closed — an unknown capability or category resolves to dangling so a
// profile that selects it is rejected rather than silently accepted.
if (resolveCapability("executor-class", "__unshipped__") !== "dangling") {
  failures.push(
    "resolveCapability did not fail closed on an unknown capability",
  );
}
if (resolveCapability("__unknown-category__", "x") !== "dangling") {
  failures.push("resolveCapability did not fail closed on an unknown category");
}

// Mutation guards — each must report a failure; a dud (empty) means a comparison
// silently broke.
const setGuards = (expected: readonly string[]): string[][] => {
  const owner = [...expected];
  return [
    compareSet("guard", [...owner, "__invented__"], owner),
    compareSet(
      "guard",
      owner.map((value, index) => (index === 0 ? `${value}__renamed` : value)),
      owner,
    ),
    compareSet("guard", owner.slice(1), owner),
  ];
};

const guards: string[][] = [
  ...setGuards(expectedVocabulary.executorClasses),
  ...setGuards(expectedVocabulary.moduleClassifications),
  ...setGuards(capabilityNames("method")),
  [entryFailure({category: "guard", name: "g", status: "maybe"}) ?? ""].filter(
    (message) => message.length > 0,
  ),
  [
    entryFailure({category: "guard", name: "g", status: "shipped"}) ?? "",
  ].filter((message) => message.length > 0),
  [
    entryFailure({category: "guard", name: "g", status: "adopter-supplied"}) ??
      "",
  ].filter((message) => message.length > 0),
];

const dudGuards = guards.filter((guardFailures) => guardFailures.length === 0);

const entryCount = entries.length;
const categoryCount = Object.keys(capabilityRegistry.categories).length;

if (dudGuards.length > 0) {
  console.error(
    `Capability registry mutation guards failed: ${dudGuards.length} tampered checks passed validation`,
  );
  process.exitCode = 1;
} else if (failures.length > 0) {
  console.error("Capability registry drift detected:");
  for (const failure of failures) {
    console.error(`- ${failure}`);
  }
  process.exitCode = 1;
} else {
  console.log(
    `Capability registry validation passed: ${entryCount} capabilities across ${categoryCount} categories classified; ${guards.length} mutation guards`,
  );
}
