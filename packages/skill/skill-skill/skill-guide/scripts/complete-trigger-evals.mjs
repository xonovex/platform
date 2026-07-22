#!/usr/bin/env node
import {
  existsSync,
  readdirSync,
  readFileSync,
  statSync,
  writeFileSync,
} from "node:fs";
import {basename, join, resolve} from "node:path";
import {fileURLToPath} from "node:url";

const MIN_PER_POLARITY = 8;
const FRONTMATTER_RE = /^---\n([\s\S]*?)\n---\n([\s\S]*)$/;
const GENERATED_QUERY_RES = [
  /^Help me handle .+ correctly in this .+ task, including the edge cases that usually get missed\.$/,
  /^I'm reviewing `[^`]+` in an? .+ project\. The happy path works, but .+ is unclear\./,
  /^A teammate says our .+ change may mishandle .+ before release\. Inspect the likely risk around /,
  /^CI started failing after we changed .+ in an? .+ project\. Use /,
  /^A review comment on `work\//,
  /^quick pre-merge sanity check: after the /,
  /^The clean Linux CI job fails only for the minimal /,
];
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
    "claude-code-guide",
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
      "claude-code-guide",
      "copilot-guide",
      "kiro-guide",
      "opencode-guide",
      "pi-guide",
    ],
  ],
  [
    "copilot-guide",
    [
      "claude-code-guide",
      "codex-guide",
      "kiro-guide",
      "opencode-guide",
      "pi-guide",
    ],
  ],
  [
    "kiro-guide",
    [
      "claude-code-guide",
      "codex-guide",
      "copilot-guide",
      "opencode-guide",
      "pi-guide",
    ],
  ],
  [
    "opencode-guide",
    [
      "claude-code-guide",
      "codex-guide",
      "copilot-guide",
      "kiro-guide",
      "pi-guide",
    ],
  ],
  [
    "pi-guide",
    [
      "claude-code-guide",
      "codex-guide",
      "copilot-guide",
      "kiro-guide",
      "opencode-guide",
    ],
  ],
]);

const SKILL_FAMILIES = [
  [
    "accessibility-guide",
    "android-analytics-guide",
    "android-wcag-guide",
    "figma-guide",
  ],
  ["android-analytics-guide", "ai-governance-guide", "datadog-guide"],
  [
    "adr-guide",
    "ddd-guide",
    "plan-guide",
    "user-stories-guide",
    "workflow-guide",
  ],
  [
    "accessibility-guide",
    "agent-governance-guide",
    "ai-governance-guide",
    "reliability-guide",
    "security-assurance-guide",
    "workflow-guide",
  ],
  [
    "atlassian-guide",
    "azure-devops-guide",
    "bitbucket-guide",
    "figma-guide",
    "github-guide",
    "gitlab-guide",
  ],
  [
    "aws-guide",
    "azure-devops-guide",
    "bitbucket-guide",
    "bitrise-guide",
    "datadog-guide",
    "docker-guide",
    "kubernetes-guide",
    "reliability-guide",
    "security-assurance-guide",
    "terraform-guide",
  ],
  [
    "bdd-guide",
    "fdd-guide",
    "tdd-guide",
    "testing-guide",
    "user-stories-guide",
    "vitest-guide",
  ],
  [
    "code-quality-guide",
    "code-review-guide",
    "github-guide",
    "gitlab-guide",
    "pull-request-guide",
    "security-assurance-guide",
  ],
  [
    "command-guide",
    "content-guide",
    "instruction-guide",
    "llmstxt-guide",
    "presentation-guide",
    "reflect-guide",
    "skill-guide",
  ],
  [
    "c99-guide",
    "lua-guide",
    "python-guide",
    "shell-scripting-guide",
    "sql-postgresql-guide",
    "typescript-guide",
  ],
  [
    "c99-game-opinionated-guide",
    "c99-guide",
    "c99-opinionated-guide",
    "cmake-guide",
    "cross-platform-guide",
    "data-oriented-design-guide",
    "debugging-guide",
    "lock-free-guide",
    "memory-management-guide",
  ],
  [
    "data-oriented-design-guide",
    "lua-guide",
    "lua-opinionated-guide",
    "typescript-to-lua-guide",
  ],
  ["typescript-guide", "typescript-to-lua-guide", "vitest-guide", "zod-guide"],
  [
    "expressjs-guide",
    "hono-guide",
    "hono-opinionated-guide",
    "typescript-guide",
    "vitest-guide",
    "zod-guide",
  ],
  [
    "astro-guide",
    "motion-guide",
    "presentation-guide",
    "react-guide",
    "remotion-guide",
    "threejs-guide",
  ],
  [
    "asset-pipeline-guide",
    "audio-guide",
    "c99-game-opinionated-guide",
    "cross-platform-guide",
    "data-model-guide",
    "data-oriented-design-guide",
    "ecs-guide",
    "editor-viewport-guide",
    "game-networking-guide",
    "gpu-rendering-guide",
    "gpu-rendering-vulkan-guide",
    "imgui-guide",
    "memory-management-guide",
    "microkernel-pattern-guide",
    "node-graph-guide",
    "threejs-guide",
  ],
  [
    "connascence-guide",
    "data-model-guide",
    "data-oriented-design-guide",
    "ddd-guide",
    "fp-guide",
    "hexagonal-pattern-guide",
    "microkernel-pattern-guide",
    "oop-guide",
    "orthogonal-pattern-guide",
  ],
  [
    "bitbucket-guide",
    "code-review-guide",
    "git-guide",
    "github-guide",
    "gitlab-guide",
    "pull-request-guide",
    "versioning-guide",
  ],
  [
    "bitrise-guide",
    "cmake-guide",
    "docker-guide",
    "kubernetes-guide",
    "moon-guide",
    "npm-guide",
    "shell-scripting-guide",
    "terraform-guide",
    "versioning-guide",
  ],
  ["content-guide", "reflect-guide"],
  ["audio-guide", "motion-guide", "remotion-guide", "strudel-guide"],
  ["data-model-guide", "node-graph-guide", "sql-postgresql-guide"],
  ["npm-guide", "pull-request-guide", "versioning-guide"],
];

const FAMILY_NAMES_BY_SKILL = new Map();
for (const [familyIndex, family] of SKILL_FAMILIES.entries()) {
  for (const skillName of family) {
    const familyNames = FAMILY_NAMES_BY_SKILL.get(skillName) ?? new Set();
    familyNames.add(familyIndex);
    FAMILY_NAMES_BY_SKILL.set(skillName, familyNames);
  }
}

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

const isGeneratedQuery = (query) =>
  GENERATED_QUERY_RES.some((pattern) => pattern.test(query));

const deduplicate = (entries) => {
  const seen = new Set();
  return entries.filter((entry) => {
    if (!isRecord(entry) || typeof entry.query !== "string") return true;
    if (isGeneratedQuery(entry.query.trim())) return false;
    if (
      entry.should_trigger === false &&
      typeof entry.rationale === "string" &&
      entry.rationale.startsWith("near miss owned by ")
    ) {
      return false;
    }
    const key = entry.query.trim().toLowerCase();
    if (seen.has(key)) return false;
    seen.add(key);
    return true;
  });
};

export const roundRobinCandidates = (siblings) => {
  const candidates = [];
  const maximumQueries = Math.max(
    0,
    ...siblings.map((sibling) => sibling.queries.length),
  );
  for (let queryIndex = 0; queryIndex < maximumQueries; queryIndex += 1) {
    for (const sibling of siblings) {
      const query = sibling.queries[queryIndex];
      if (typeof query === "string") {
        candidates.push({owner: sibling.name, query});
      }
    }
  }
  return candidates;
};

const trainCountFor = (entryCount) =>
  Math.max(2, Math.min(entryCount - 2, Math.ceil(entryCount * 0.6)));

export const alternateFutureSplits = (queries) => {
  const trainCount = trainCountFor(queries.length);
  const train = queries.slice(0, trainCount);
  const validation = queries.slice(trainCount);
  const ordered = [];
  const maximum = Math.max(train.length, validation.length);
  for (let index = 0; index < maximum; index += 1) {
    const trainQuery = train[index];
    if (typeof trainQuery === "string") ordered.push(trainQuery);
    const validationQuery = validation[index];
    if (typeof validationQuery === "string") ordered.push(validationQuery);
  }
  return ordered;
};

const addPositiveSeeds = (skill) => {
  const positives = skill.queries.filter(
    (entry) => entry.should_trigger === true,
  );
  if (positives.length < MIN_PER_POLARITY) {
    throw new Error(
      `${skill.name} has only ${String(positives.length)} curated positive seeds; add realistic queries before completing the catalog`,
    );
  }
};

const familyOverlap = (leftName, rightName) => {
  const leftFamilies = FAMILY_NAMES_BY_SKILL.get(leftName) ?? new Set();
  const rightFamilies = FAMILY_NAMES_BY_SKILL.get(rightName) ?? new Set();
  return [...leftFamilies].filter((family) => rightFamilies.has(family)).length;
};

const namedReferences = (skill, skills) =>
  new Set(
    skills
      .filter(
        (candidate) =>
          candidate.name !== skill.name && skill.body.includes(candidate.name),
      )
      .map((candidate) => candidate.name),
  );

export const selectSiblingNames = (skill, skills) => {
  const explicit = HARNESS_SIBLINGS.get(skill.name);
  if (explicit !== undefined) return explicit;
  const ownWords = wordSet(skill.description);
  const outgoing = namedReferences(skill, skills);
  const incoming = new Set(
    skills
      .filter(
        (candidate) =>
          candidate.name !== skill.name && candidate.body.includes(skill.name),
      )
      .map((candidate) => candidate.name),
  );
  const ranked = skills
    .filter((candidate) => candidate.name !== skill.name)
    .map((candidate) => ({
      name: candidate.name,
      affinity:
        (outgoing.has(candidate.name) ? 10_000 : 0) +
        (incoming.has(candidate.name) ? 5_000 : 0) +
        familyOverlap(skill.name, candidate.name) * 100,
      lexical: similarity(ownWords, wordSet(candidate.description)),
    }))
    .toSorted(
      (left, right) =>
        right.affinity - left.affinity ||
        right.lexical - left.lexical ||
        left.name.localeCompare(right.name),
    );
  const related = ranked.filter((candidate) => candidate.affinity > 0);
  return (related.length > 0 ? related : ranked)
    .slice(0, 4)
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
  const siblings = selectSiblingNames(skill, allSkills).flatMap(
    (siblingName) => {
      const sibling = skillsByName.get(siblingName);
      if (sibling === undefined) return [];
      return [
        {
          name: sibling.name,
          queries: alternateFutureSplits(
            sibling.queries
              .filter((query) => query.should_trigger === true)
              .map((entry) => entry.query)
              .filter((query) => typeof query === "string"),
          ),
        },
      ];
    },
  );
  const candidates = roundRobinCandidates(siblings).map(({owner, query}) => ({
    query,
    rationale: `near miss owned by ${owner}`,
  }));
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
    const trainCount = trainCountFor(entries.length);
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

if (
  process.argv[1] !== undefined &&
  resolve(process.argv[1]) === fileURLToPath(import.meta.url)
) {
  main();
}
