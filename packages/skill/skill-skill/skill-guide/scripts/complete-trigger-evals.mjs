#!/usr/bin/env node
import {
  existsSync,
  readdirSync,
  readFileSync,
  statSync,
  writeFileSync,
} from "node:fs";
import {basename, join, resolve} from "node:path";

const MIN_PER_POLARITY = 8;
const FRONTMATTER_RE = /^---\n([\s\S]*?)\n---\n([\s\S]*)$/;
const REFERENCE_RE = /references\/([a-z0-9][a-z0-9-]*\.md)/g;
const WORD_RE = /[a-z][a-z0-9-]{2,}/g;
const STOP_WORDS = new Set([
  "about",
  "apply",
  "editing",
  "even",
  "files",
  "guide",
  "prompts",
  "say",
  "skill",
  "this",
  "triggers",
  "user",
  "when",
  "with",
  "without",
  "writing",
]);

const HARNESS_SIBLINGS = new Map([
  [
    "code-harness-guide",
    [
      "codex-guide",
      "copilot-guide",
      "kiro-guide",
      "opencode-guide",
      "pi-guide",
    ],
  ],
  [
    "codex-guide",
    [
      "code-harness-guide",
      "copilot-guide",
      "kiro-guide",
      "opencode-guide",
      "pi-guide",
    ],
  ],
  [
    "copilot-guide",
    [
      "code-harness-guide",
      "codex-guide",
      "kiro-guide",
      "opencode-guide",
      "pi-guide",
    ],
  ],
  [
    "kiro-guide",
    [
      "code-harness-guide",
      "codex-guide",
      "copilot-guide",
      "opencode-guide",
      "pi-guide",
    ],
  ],
  [
    "opencode-guide",
    [
      "code-harness-guide",
      "codex-guide",
      "copilot-guide",
      "kiro-guide",
      "pi-guide",
    ],
  ],
  [
    "pi-guide",
    [
      "code-harness-guide",
      "codex-guide",
      "copilot-guide",
      "kiro-guide",
      "opencode-guide",
    ],
  ],
]);

const isRecord = (value) =>
  typeof value === "object" && value !== null && !Array.isArray(value);

const loadJson = (path) => JSON.parse(readFileSync(path, "utf8"));

const guideDirectories = (root) => {
  const found = [];
  for (const packageName of readdirSync(root).toSorted()) {
    const packageDirectory = join(root, packageName);
    if (
      !packageName.startsWith("skill-") ||
      !statSync(packageDirectory).isDirectory()
    ) {
      continue;
    }
    for (const guideName of readdirSync(packageDirectory).toSorted()) {
      const guideDirectory = join(packageDirectory, guideName);
      if (
        statSync(guideDirectory).isDirectory() &&
        existsSync(join(guideDirectory, "SKILL.md")) &&
        existsSync(join(guideDirectory, "evals.json"))
      ) {
        found.push(guideDirectory);
      }
    }
  }
  return found;
};

const parseSkill = (guideDirectory) => {
  const text = readFileSync(join(guideDirectory, "SKILL.md"), "utf8");
  const match = FRONTMATTER_RE.exec(text);
  if (!match)
    throw new Error(`invalid SKILL.md frontmatter: ${guideDirectory}`);
  const rawFrontmatter = match[1] ?? "";
  const name = /^name:\s*([^\s]+)\s*$/m.exec(rawFrontmatter)?.[1];
  const rawDescription = /^description:\s*(.+)\s*$/m.exec(rawFrontmatter)?.[1];
  if (name === undefined || rawDescription === undefined) {
    throw new Error(`invalid skill name: ${guideDirectory}`);
  }
  const description = rawDescription.startsWith('"')
    ? JSON.parse(rawDescription)
    : rawDescription;
  if (typeof description !== "string") {
    throw new Error(`invalid skill description: ${guideDirectory}`);
  }
  return {
    body: match[2] ?? "",
    description,
    name,
  };
};

const wordSet = (text) =>
  new Set(
    (text.toLowerCase().match(WORD_RE) ?? []).filter(
      (word) => !STOP_WORDS.has(word),
    ),
  );

const similarity = (left, right) => {
  const intersection = [...left].filter((word) => right.has(word)).length;
  const union = new Set([...left, ...right]).size;
  return union === 0 ? 0 : intersection / union;
};

const humanize = (value) =>
  value
    .replace(/-guide$/, "")
    .replaceAll("-", " ")
    .replace(/\b\w/g, (letter) => letter.toUpperCase());

const referenceTopics = (body) => {
  const topics = [];
  for (const match of body.matchAll(REFERENCE_RE)) {
    const file = match[1];
    if (file !== undefined && !topics.includes(file)) topics.push(file);
  }
  return topics.map((file) => file.replace(/\.md$/, "").replaceAll("-", " "));
};

const headingTopics = (body) =>
  [...body.matchAll(/^##\s+(.+\S)\s*$/gm)]
    .map((match) => match[1])
    .filter(
      (heading) =>
        typeof heading === "string" &&
        !/^(gotchas|progressive disclosure|example|quick contrast)$/i.test(
          heading,
        ),
    );

const descriptionTopics = (description) => {
  const triggerText = /Triggers on (.*?)(?:,? even when|\.(?:\s|$))/i.exec(
    description,
  )?.[1];
  if (triggerText === undefined) return [];
  return triggerText
    .split(/,|\bor\b|\band\b/)
    .map((topic) => topic.trim().replace(/^prompts about\s+/i, ""))
    .filter((topic) => topic.length > 3);
};

const deduplicate = (entries) => {
  const seen = new Set();
  return entries.filter((entry) => {
    if (!isRecord(entry) || typeof entry.query !== "string") return true;
    const key = entry.query.trim().toLowerCase();
    if (seen.has(key)) return false;
    seen.add(key);
    return true;
  });
};

const addPositiveSeeds = (skill) => {
  const positives = skill.queries.filter(
    (entry) => entry.should_trigger === true,
  );
  const seen = new Set(
    skill.queries
      .filter((entry) => typeof entry.query === "string")
      .map((entry) => entry.query.trim().toLowerCase()),
  );
  const candidates = [
    ...skill.outputPrompts.map((query) => ({
      query,
      rationale:
        "output-eval task also exercises this skill's routing boundary",
    })),
    ...referenceTopics(skill.body).map((topic) => ({
      query: `Help me handle ${topic} correctly in this ${humanize(skill.name)} task, including the edge cases that usually get missed.`,
      rationale: `implicit task cue derived from the ${topic} reference`,
    })),
    ...headingTopics(skill.body).map((topic) => ({
      query: `Review this ${humanize(skill.name)} change for ${topic.toLowerCase()} and fix the concrete issues you find.`,
      rationale: `implicit task cue derived from the ${topic} section`,
    })),
    ...descriptionTopics(skill.description).map((topic) => ({
      query: `Help me with ${topic}; apply the relevant ${humanize(skill.name)} conventions and call out the important edge cases.`,
      rationale: `routing cue derived from the description's ${topic} trigger`,
    })),
  ];
  for (const candidate of candidates) {
    if (positives.length >= MIN_PER_POLARITY) break;
    const key = candidate.query.trim().toLowerCase();
    if (seen.has(key)) continue;
    const entry = {...candidate, should_trigger: true};
    skill.queries.push(entry);
    positives.push(entry);
    seen.add(key);
  }
  if (positives.length < MIN_PER_POLARITY) {
    throw new Error(
      `${skill.name} has only ${String(positives.length)} usable positive seeds`,
    );
  }
};

const siblingNames = (skill, skills) => {
  const explicit = HARNESS_SIBLINGS.get(skill.name);
  if (explicit !== undefined) return explicit;
  const ownWords = wordSet(skill.description);
  return skills
    .filter((candidate) => candidate.name !== skill.name)
    .map((candidate) => ({
      name: candidate.name,
      score: similarity(ownWords, wordSet(candidate.description)),
    }))
    .toSorted(
      (left, right) =>
        right.score - left.score || left.name.localeCompare(right.name),
    )
    .slice(0, 8)
    .map((candidate) => candidate.name);
};

const addNegativeSeeds = (skill, skillsByName, allSkills) => {
  const negatives = skill.queries.filter(
    (entry) => entry.should_trigger === false,
  );
  const seen = new Set(
    skill.queries
      .filter((entry) => typeof entry.query === "string")
      .map((entry) => entry.query.trim().toLowerCase()),
  );
  const candidates = [];
  for (const siblingName of siblingNames(skill, allSkills)) {
    const sibling = skillsByName.get(siblingName);
    if (sibling === undefined) continue;
    for (const entry of sibling.queries.filter(
      (query) => query.should_trigger === true,
    )) {
      if (typeof entry.query !== "string") continue;
      candidates.push({
        query: entry.query,
        rationale: `near miss owned by ${sibling.name}`,
      });
    }
  }
  for (const candidate of candidates) {
    if (negatives.length >= MIN_PER_POLARITY) break;
    const key = candidate.query.trim().toLowerCase();
    if (seen.has(key)) continue;
    const entry = {...candidate, should_trigger: false};
    skill.queries.push(entry);
    negatives.push(entry);
    seen.add(key);
  }
  if (negatives.length < MIN_PER_POLARITY) {
    throw new Error(
      `${skill.name} has only ${String(negatives.length)} usable negative seeds`,
    );
  }
};

const assignSplits = (queries) => {
  for (const polarity of [true, false]) {
    const entries = queries.filter(
      (entry) => entry.should_trigger === polarity,
    );
    const trainCount = Math.max(
      2,
      Math.min(entries.length - 2, Math.ceil(entries.length * 0.6)),
    );
    for (const [index, entry] of entries.entries()) {
      entry.split = index < trainCount ? "train" : "validation";
    }
  }
};

const main = () => {
  const catalogRoot = resolve(process.argv[2] ?? "packages/skill");
  const skills = guideDirectories(catalogRoot).map((guideDirectory) => {
    const parsed = parseSkill(guideDirectory);
    const output = loadJson(join(guideDirectory, "evals.json"));
    if (!isRecord(output) || !Array.isArray(output.evals)) {
      throw new Error(`invalid evals.json: ${guideDirectory}`);
    }
    const queryPath = join(guideDirectory, "eval-queries.json");
    const queries = existsSync(queryPath) ? loadJson(queryPath) : [];
    if (!Array.isArray(queries)) {
      throw new Error(`invalid eval-queries.json: ${guideDirectory}`);
    }
    return {
      ...parsed,
      guideDirectory,
      outputPrompts: output.evals
        .filter(isRecord)
        .map((entry) => entry.prompt)
        .filter((prompt) => typeof prompt === "string"),
      queries: deduplicate(queries),
    };
  });

  for (const skill of skills) addPositiveSeeds(skill);
  const skillsByName = new Map(skills.map((skill) => [skill.name, skill]));
  for (const skill of skills) addNegativeSeeds(skill, skillsByName, skills);

  for (const skill of skills) {
    assignSplits(skill.queries);
    const path = join(skill.guideDirectory, "eval-queries.json");
    writeFileSync(path, JSON.stringify(skill.queries, null, 2) + "\n");
    const positive = skill.queries.filter(
      (entry) => entry.should_trigger === true,
    ).length;
    const negative = skill.queries.filter(
      (entry) => entry.should_trigger === false,
    ).length;
    process.stdout.write(
      `${basename(skill.guideDirectory)}\t${String(positive)}\t${String(negative)}\n`,
    );
  }
};

main();
