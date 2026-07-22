import {readdirSync, readFileSync} from "node:fs";
import {join, resolve} from "node:path";
import {
  isDirectory,
  isFile,
  resolveGuideDirectory,
} from "@xonovex/script-moon-common/fs";
import {parseFrontmatterName} from "./cli.js";
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

const readQueries = (path: string): readonly Query[] => {
  let input: unknown;
  try {
    input = JSON.parse(readFileSync(path, "utf8"));
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

const catalogCandidates = (catalogRoot: string): readonly RoutingCandidate[] =>
  readdirSync(catalogRoot)
    .filter((entry) => entry.startsWith("skill-"))
    .map((entry) => join(catalogRoot, entry))
    .filter(isDirectory)
    .flatMap((pluginDirectory) => {
      const guideDirectory = resolveGuideDirectory(pluginDirectory);
      const skillFile = join(guideDirectory, "SKILL.md");
      const queriesFile = join(guideDirectory, "eval-queries.json");
      if (!isFile(skillFile) || !isFile(queriesFile)) return [];
      const shortName = parseFrontmatterName(skillFile);
      if (shortName === undefined) {
        throw new Error(`skill name missing from ${skillFile}`);
      }
      return [{guideDirectory, pluginDirectory, shortName}];
    });

export const buildRoutingScenarios = (
  catalogRoot: string,
): readonly RoutingScenario[] => {
  const root = resolve(catalogRoot);
  if (!isDirectory(root)) throw new Error(`catalog root not found: ${root}`);

  const occurrencesByQuery = new Map<string, QueryOccurrence[]>();
  for (const candidate of catalogCandidates(root)) {
    const queries = readQueries(
      join(candidate.guideDirectory, "eval-queries.json"),
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

export const missingValidationRoutingOwners = (
  catalogRoot: string,
): readonly string[] => {
  const scenarios = buildRoutingScenarios(catalogRoot);
  const validationOwners = new Set(
    scenarios
      .filter(({split}) => split === "validation")
      .map(({expectedSkill}) => expectedSkill),
  );
  return catalogCandidates(resolve(catalogRoot))
    .map(({shortName}) => shortName)
    .filter((shortName) => !validationOwners.has(shortName))
    .toSorted();
};
