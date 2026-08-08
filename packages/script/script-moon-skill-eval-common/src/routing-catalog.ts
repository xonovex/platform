import {join, resolve} from "node:path";
import {
  nodeFileSystem,
  type FileSystem,
} from "@xonovex/script-moon-common/file-system";
import {parseFrontmatterName} from "@xonovex/script-moon-common/frontmatter";
import {
  isDirectory,
  isFile,
  resolveGuideDirectory,
} from "@xonovex/script-moon-common/fs";
import {parseQueries, type Query} from "./validation.js";

export interface RoutingCandidate {
  readonly guideDirectory: string;
  readonly pluginDirectory: string;
  readonly shortName: string;
}

export interface RoutingScenario {
  readonly candidates: readonly RoutingCandidate[];
  readonly expectedSkill: string;
  readonly query: string;
  readonly split: "train" | "validation";
}

interface QueryOccurrence {
  readonly candidate: RoutingCandidate;
  readonly query: Query;
}

const readQueries = (path: string, fs: FileSystem): readonly Query[] => {
  let input: unknown;
  try {
    input = JSON.parse(fs.readText(path));
  } catch (error: unknown) {
    const detail = error instanceof Error ? error.message : String(error);
    throw new Error(`invalid JSON in ${path}: ${detail}`);
  }
  const result = parseQueries(input);
  if (!result.success) {
    throw new Error(`invalid queries in ${path}: ${result.error}`);
  }
  return result.data;
};

const catalogCandidates = (
  catalogRoot: string,
  fs: FileSystem,
): readonly RoutingCandidate[] =>
  fs
    .readDirectory(catalogRoot)
    .filter((entry) => entry.startsWith("skill-"))
    .map((entry) => join(catalogRoot, entry))
    .filter((path) => isDirectory(path, fs))
    .flatMap((pluginDirectory) => {
      const guideDirectory = resolveGuideDirectory(pluginDirectory, fs);
      const skillFile = join(guideDirectory, "SKILL.md");
      const queriesFile = join(guideDirectory, "eval-queries.json");
      if (!isFile(skillFile, fs) || !isFile(queriesFile, fs)) return [];
      const shortName = parseFrontmatterName(skillFile, fs);
      if (shortName === undefined) {
        throw new Error(`skill name missing from ${skillFile}`);
      }
      return [{guideDirectory, pluginDirectory, shortName}];
    });

export const buildRoutingScenarios = (
  catalogRoot: string,
  fs: FileSystem = nodeFileSystem,
): readonly RoutingScenario[] => {
  const root = resolve(catalogRoot);
  if (!isDirectory(root, fs))
    throw new Error(`catalog root not found: ${root}`);

  const occurrencesByQuery = new Map<string, QueryOccurrence[]>();
  for (const candidate of catalogCandidates(root, fs)) {
    const queries = readQueries(
      join(candidate.guideDirectory, "eval-queries.json"),
      fs,
    );
    for (const query of queries) {
      const occurrences = occurrencesByQuery.get(query.query) ?? [];
      occurrences.push({candidate, query});
      occurrencesByQuery.set(query.query, occurrences);
    }
  }

  const scenarios: RoutingScenario[] = [];
  for (const [queryText, occurrences] of occurrencesByQuery) {
    const positives = occurrences.filter(({query}) => query.should_trigger);
    const negatives = occurrences.filter(({query}) => !query.should_trigger);
    if (positives.length !== 1 || negatives.length === 0) continue;
    const positive = positives[0];
    if (positive === undefined) continue;
    const candidateByName = new Map<string, RoutingCandidate>();
    for (const occurrence of occurrences) {
      candidateByName.set(occurrence.candidate.shortName, occurrence.candidate);
    }
    scenarios.push({
      candidates: [...candidateByName.values()].toSorted((a, b) =>
        a.shortName.localeCompare(b.shortName),
      ),
      expectedSkill: positive.candidate.shortName,
      query: queryText,
      split: positive.query.split ?? "train",
    });
  }
  return scenarios.toSorted(
    (a, b) =>
      a.expectedSkill.localeCompare(b.expectedSkill) ||
      a.query.localeCompare(b.query),
  );
};

// The skill that claims each query, keyed by query text. A query only one skill
// declares positive has a single owner; one no skill declares positive is absent
// from the map. Callers use it to tell a query the catalog assigns elsewhere from
// one that belongs to no skill at all.
export const catalogQueryOwners = (
  catalogRoot: string,
  fs: FileSystem = nodeFileSystem,
): ReadonlyMap<string, string> => {
  const root = resolve(catalogRoot);
  if (!isDirectory(root, fs))
    throw new Error(`catalog root not found: ${root}`);

  const positivesByQuery = new Map<string, string[]>();
  for (const candidate of catalogCandidates(root, fs)) {
    const queries = readQueries(
      join(candidate.guideDirectory, "eval-queries.json"),
      fs,
    );
    for (const query of queries) {
      if (!query.should_trigger) continue;
      const owners = positivesByQuery.get(query.query) ?? [];
      owners.push(candidate.shortName);
      positivesByQuery.set(query.query, owners);
    }
  }

  const owners = new Map<string, string>();
  for (const [queryText, names] of positivesByQuery) {
    const only = names.length === 1 ? names[0] : undefined;
    if (only !== undefined) owners.set(queryText, only);
  }
  return owners;
};

export interface QueryOwnerConflict {
  readonly owners: readonly string[];
  readonly query: string;
}

// normalizeQuery folds the differences that make a copied query look distinct:
// case, run-together whitespace, and closing punctuation. Two skills claiming
// the same normalized text are claiming the same route.
const normalizeQuery = (query: string): string =>
  query
    .toLowerCase()
    .split(/\s+/u)
    .filter(Boolean)
    .join(" ")
    .replace(/[.?!]+$/u, "");

// Queries more than one skill declares positive. buildRoutingScenarios and
// catalogQueryOwners both drop these silently, so an undetected conflict costs
// the query its routing coverage on top of the ambiguity itself.
export const conflictingQueryOwners = (
  catalogRoot: string,
  fs: FileSystem = nodeFileSystem,
): readonly QueryOwnerConflict[] => {
  const root = resolve(catalogRoot);
  if (!isDirectory(root, fs))
    throw new Error(`catalog root not found: ${root}`);

  const claimants = new Map<string, {owners: Set<string>; query: string}>();
  for (const candidate of catalogCandidates(root, fs)) {
    const queries = readQueries(
      join(candidate.guideDirectory, "eval-queries.json"),
      fs,
    );
    for (const query of queries) {
      if (!query.should_trigger) continue;
      const key = normalizeQuery(query.query);
      const entry = claimants.get(key) ?? {
        owners: new Set(),
        query: query.query,
      };
      entry.owners.add(candidate.shortName);
      claimants.set(key, entry);
    }
  }

  return [...claimants.values()]
    .filter(({owners}) => owners.size > 1)
    .map(({owners, query}) => ({owners: [...owners].toSorted(), query}))
    .toSorted((a, b) => a.query.localeCompare(b.query));
};

export interface UnresolvedOperation {
  readonly operation: string;
  readonly rationale: string;
  readonly skill: string;
}

const OPERATION_MENTION = /`?([a-z0-9]+(?:-[a-z0-9]+)+)`?\s+operation/gu;

const guideTopic = (shortName: string): string =>
  shortName.replace(/-guide$/u, "");

// Rationales that name an operation of a catalog skill which owns no reference
// file for it. Scoped to tokens prefixed with a real skill topic so that plain
// English compounds ('a version-control operation') are not identifiers.
export const unresolvedOperationRationales = (
  catalogRoot: string,
  fs: FileSystem = nodeFileSystem,
): readonly UnresolvedOperation[] => {
  const root = resolve(catalogRoot);
  if (!isDirectory(root, fs))
    throw new Error(`catalog root not found: ${root}`);

  const candidates = catalogCandidates(root, fs);
  const topics = candidates.map(({shortName}) => guideTopic(shortName));
  const references = new Set(
    candidates.flatMap(({guideDirectory}) => {
      const directory = join(guideDirectory, "references");
      if (!isDirectory(directory, fs)) return [];
      return fs
        .readDirectory(directory)
        .filter((entry) => entry.endsWith(".md"))
        .map((entry) => entry.slice(0, -".md".length));
    }),
  );

  const findings: UnresolvedOperation[] = [];
  for (const candidate of candidates) {
    const queries = readQueries(
      join(candidate.guideDirectory, "eval-queries.json"),
      fs,
    );
    for (const {rationale} of queries) {
      for (const [, operation] of rationale.matchAll(OPERATION_MENTION)) {
        if (operation === undefined || references.has(operation)) continue;
        if (topics.every((topic) => !operation.startsWith(`${topic}-`)))
          continue;
        findings.push({
          operation,
          rationale,
          skill: candidate.shortName,
        });
      }
    }
  }
  return findings.toSorted(
    (a, b) =>
      a.skill.localeCompare(b.skill) || a.operation.localeCompare(b.operation),
  );
};

export const missingValidationRoutingOwners = (
  catalogRoot: string,
  fs: FileSystem = nodeFileSystem,
): readonly string[] => {
  const scenarios = buildRoutingScenarios(catalogRoot, fs);
  const validationOwners = new Set(
    scenarios
      .filter(({split}) => split === "validation")
      .map(({expectedSkill}) => expectedSkill),
  );
  return catalogCandidates(resolve(catalogRoot), fs)
    .map(({shortName}) => shortName)
    .filter((shortName) => !validationOwners.has(shortName))
    .toSorted();
};
