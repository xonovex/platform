#!/usr/bin/env node
import {readdirSync, readFileSync, statSync} from "node:fs";
import {join, resolve} from "node:path";
import {fileURLToPath} from "node:url";

const isDirectory = (path) => {
  try {
    return statSync(path).isDirectory();
  } catch {
    return false;
  }
};

const isFile = (path) => {
  try {
    return statSync(path).isFile();
  } catch {
    return false;
  }
};

const guideDirectories = (catalogRoot) =>
  readdirSync(catalogRoot)
    .filter((entry) => entry.startsWith("skill-"))
    .map((entry) => join(catalogRoot, entry))
    .filter(isDirectory)
    .flatMap((packageDirectory) =>
      readdirSync(packageDirectory)
        .map((entry) => join(packageDirectory, entry))
        .filter((directory) => isFile(join(directory, "SKILL.md"))),
    )
    .toSorted();

const readJson = (path) => JSON.parse(readFileSync(path, "utf8"));

const lineCount = (path) =>
  (readFileSync(path, "utf8").match(/\n/g) ?? []).length;

const skillName = (guideDirectory) => {
  const skill = readFileSync(join(guideDirectory, "SKILL.md"), "utf8");
  const name = /^name:\s*"?([^"\n]+)"?$/m.exec(skill)?.[1]?.trim();
  if (name === undefined) {
    throw new Error(`skill name missing from ${guideDirectory}`);
  }
  return name;
};

const referenceLineCount = (guideDirectory) => {
  const references = join(guideDirectory, "references");
  if (!isDirectory(references)) return 0;
  return readdirSync(references)
    .filter((entry) => entry.endsWith(".md"))
    .map((entry) => lineCount(join(references, entry)))
    .reduce((total, lines) => total + lines, 0);
};

const outputEvidence = (guideDirectory) => {
  const output = readJson(join(guideDirectory, "evals.json"));
  if (
    typeof output !== "object" ||
    output === null ||
    !Array.isArray(output.evals) ||
    typeof output.tier !== "string"
  ) {
    throw new Error(`invalid evals.json: ${guideDirectory}`);
  }
  return {count: output.evals.length, tier: output.tier};
};

export const catalogStats = (catalogRoot) => {
  const root = resolve(catalogRoot);
  const guides = guideDirectories(root);
  const tiers = {aggressive: 0, moderate: 0, conservative: 0};
  let distilledLines = 0;
  let outputEvals = 0;
  let triggerQueries = 0;
  const occurrencesByQuery = new Map();

  for (const guideDirectory of guides) {
    const name = skillName(guideDirectory);
    distilledLines +=
      lineCount(join(guideDirectory, "SKILL.md")) +
      referenceLineCount(guideDirectory);
    const output = outputEvidence(guideDirectory);
    if (!(output.tier in tiers)) {
      throw new Error(
        `invalid output tier '${output.tier}': ${guideDirectory}`,
      );
    }
    tiers[output.tier] += 1;
    outputEvals += output.count;

    const queries = readJson(join(guideDirectory, "eval-queries.json"));
    if (!Array.isArray(queries)) {
      throw new Error(`invalid eval-queries.json: ${guideDirectory}`);
    }
    triggerQueries += queries.length;
    for (const query of queries) {
      if (
        typeof query !== "object" ||
        query === null ||
        typeof query.query !== "string" ||
        typeof query.should_trigger !== "boolean"
      ) {
        throw new Error(`invalid trigger query: ${guideDirectory}`);
      }
      const occurrences = occurrencesByQuery.get(query.query) ?? [];
      occurrences.push({
        owner: name,
        shouldTrigger: query.should_trigger,
        split: query.split === "validation" ? "validation" : "train",
      });
      occurrencesByQuery.set(query.query, occurrences);
    }
  }

  const scenarios = [...occurrencesByQuery.values()].flatMap((occurrences) => {
    const positives = occurrences.filter(({shouldTrigger}) => shouldTrigger);
    const negatives = occurrences.filter(({shouldTrigger}) => !shouldTrigger);
    return positives.length === 1 && negatives.length > 0 ? positives : [];
  });
  const train = scenarios.filter(({split}) => split === "train");
  const validation = scenarios.filter(({split}) => split === "validation");
  return {
    distilledLines,
    outputEvals,
    routing: {
      scenarios: scenarios.length,
      train: train.length,
      validation: validation.length,
      validationOwners: new Set(validation.map(({owner}) => owner)).size,
    },
    skills: guides.length,
    tiers,
    triggerQueries,
  };
};

const formatInteger = (value) => value.toLocaleString("en-US");

export const renderCatalogStats = (stats) =>
  [
    `- Current catalog: **${formatInteger(stats.skills)} skills**, **${formatInteger(stats.distilledLines)}** \`SKILL.md\` + \`references/\` lines, **${formatInteger(stats.outputEvals)} output evals**, and **${formatInteger(stats.triggerQueries)} trigger queries**.`,
    `- Tiers: **${formatInteger(stats.tiers.aggressive)} aggressive**, **${formatInteger(stats.tiers.moderate)} moderate**, **${formatInteger(stats.tiers.conservative)} conservative**.`,
    `- Competitive routing: **${formatInteger(stats.routing.scenarios)} scenarios** (**${formatInteger(stats.routing.train)} train**, **${formatInteger(stats.routing.validation)} validation**) with **${formatInteger(stats.routing.validationOwners)}/${formatInteger(stats.skills)} skills** owning at least one validation scenario.`,
  ].join("\n");

const main = () => {
  const catalogRoot = process.argv[2] ?? "packages/skill";
  const stats = catalogStats(catalogRoot);
  if (process.argv.includes("--json")) {
    process.stdout.write(`${JSON.stringify(stats, undefined, 2)}\n`);
    return;
  }
  process.stdout.write(`${renderCatalogStats(stats)}\n`);
};

if (
  process.argv[1] !== undefined &&
  resolve(process.argv[1]) === fileURLToPath(import.meta.url)
) {
  main();
}
