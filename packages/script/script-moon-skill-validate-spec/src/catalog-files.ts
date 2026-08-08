import {join} from "node:path";
import {
  nodeFileSystem,
  type FileSystem,
} from "@xonovex/script-moon-common/file-system";
import {isRecord} from "@xonovex/script-moon-common/records";

const OUTPUT_TIERS = new Set(["aggressive", "moderate", "conservative"]);
const QUERY_SPLITS = new Set(["train", "validation"]);
const URL_FIELD_RE = /\*\*URLs?:\*\*/i;
const HTTP_URL_RE = /https?:\/\/\S+/i;
const PROVENANCE_RE = /\*\*Provenance:\*\*\s*\S+/i;
const REVIEWED_RE = /\*\*Last reviewed:\*\*\s*\d{4}-\d{2}-\d{2}/i;
const VERSION_FIELD_RE = /\*\*Version:\*\*\s*\S+/i;
const CONTENT_SHA256_FIELD_RE = /\*\*Content SHA256:\*\*\s*[0-9a-f]{64}/i;
const CHECKOUT_FIELD_RE = /\*\*Checkout:\*\*\s*\S+/i;
const COMMIT_FIELD_RE = /\*\*Commit:\*\*\s*[0-9a-f]{7,40}/i;
const WATCH_FIELD_RE = /\*\*Watch:\*\*\s*\S+/i;
const VERSIONED_DESCRIPTION_RE = /\b\d+(?:\.\d+){0,2}\+/;
const UNSAFE_CREDENTIAL_PATTERNS: readonly [RegExp, string][] = [
  [
    /\becho\s+(?:["']?\$[A-Z_]*(?:TOKEN|KEY|PAT|SECRET|PASSWORD)["']?|<token>)\s*\|/i,
    "pipes a secret through echo",
  ],
  [/<\s*token\.txt\b/i, "reads a secret from token.txt"],
  [
    /\bexport\s+[A-Z_]*(?:TOKEN|KEY|PAT|SECRET|PASSWORD)\s*=\s*(?!["']?\$)\S+/i,
    "assigns a secret in an export command",
  ],
];
const GENERIC_TRIGGER_QUERY_RES = [
  /^Help me handle .+ correctly in this .+ task, including the edge cases that usually get missed\.$/i,
  /^I'm reviewing `[^`]+` in an? .+ project\. The happy path works, but .+ is unclear\./i,
  /^A teammate says our .+ change may mishandle .+ before release\. Inspect the likely risk around /i,
  /^CI started failing after we changed .+ in an? .+ project\. Use /i,
  /^A review comment on `work\//i,
  /^quick pre-merge sanity check: after the /i,
  /^The clean Linux CI job fails only for the minimal /i,
];

export interface CatalogFileReport {
  readonly passes: readonly string[];
  readonly errors: readonly string[];
}

const readJson = (path: string, errors: string[], fs: FileSystem): unknown => {
  try {
    return JSON.parse(fs.readText(path)) as unknown;
  } catch (error) {
    errors.push(
      `catalog: ${path.split("/").at(-1) ?? path} is not valid JSON: ${error instanceof Error ? error.message : String(error)}`,
    );
    return undefined;
  }
};

const isGenericTriggerQuery = (query: string): boolean =>
  GENERIC_TRIGGER_QUERY_RES.some((pattern) => pattern.test(query.trim()));

interface TriggerEntry {
  readonly generatedNegativeOwner: string | undefined;
  readonly shouldTrigger: boolean;
  readonly split: string | undefined;
}

const checkTriggerQuery = (
  query: unknown,
  index: number,
  seen: Set<string>,
  errors: string[],
): void => {
  if (typeof query !== "string" || query.trim().length === 0) {
    errors.push(
      `catalog: eval-queries.json entry ${String(index + 1)} needs a non-empty query`,
    );
    return;
  }
  const key = query.trim().toLowerCase();
  if (seen.has(key)) {
    errors.push(`catalog: duplicate trigger query '${query}'`);
  }
  if (isGenericTriggerQuery(query)) {
    errors.push(
      `catalog: generic trigger eval query must be replaced: '${query}'`,
    );
  }
  seen.add(key);
};

const checkTriggerEntry = (
  value: unknown,
  index: number,
  seen: Set<string>,
  errors: string[],
): TriggerEntry | undefined => {
  if (!isRecord(value)) {
    errors.push(
      `catalog: eval-queries.json entry ${String(index + 1)} must be an object`,
    );
    return undefined;
  }
  checkTriggerQuery(value.query, index, seen, errors);
  if (typeof value.should_trigger !== "boolean") {
    errors.push(
      `catalog: eval-queries.json entry ${String(index + 1)} needs boolean should_trigger`,
    );
    return undefined;
  }
  const rationale = value.rationale;
  const generatedNegativeOwner =
    typeof rationale === "string" && !value.should_trigger
      ? /^near miss owned by (\S+)$/.exec(rationale.trim())?.[1]
      : undefined;
  const split = value.split;
  if (typeof split !== "string" || !QUERY_SPLITS.has(split)) {
    errors.push(
      `catalog: eval-queries.json entry ${String(index + 1)} needs split train or validation`,
    );
    return {
      generatedNegativeOwner,
      shouldTrigger: value.should_trigger,
      split: undefined,
    };
  }
  return {generatedNegativeOwner, shouldTrigger: value.should_trigger, split};
};

const checkTriggerMinimums = (
  positive: number,
  negative: number,
  splitCounts: ReadonlyMap<string, number>,
  errors: string[],
): void => {
  if (positive < 8 || negative < 8) {
    errors.push(
      `catalog: trigger evals need at least 8 positive and 8 negative queries (found ${String(positive)}/${String(negative)})`,
    );
  }
  for (const polarity of ["positive", "negative"]) {
    for (const split of QUERY_SPLITS) {
      if ((splitCounts.get(`${polarity}:${split}`) ?? 0) < 2) {
        errors.push(
          `catalog: trigger evals need at least 2 ${polarity} ${split} queries`,
        );
      }
    }
  }
};

const checkOutputEvals = (
  skillDir: string,
  skillName: unknown,
  passes: string[],
  errors: string[],
  fs: FileSystem,
): void => {
  const path = join(skillDir, "evals.json");
  if (!fs.isFile(path)) {
    errors.push("catalog: missing evals.json");
    return;
  }
  const raw = readJson(path, errors, fs);
  if (!isRecord(raw)) {
    if (raw !== undefined) errors.push("catalog: evals.json must be an object");
    return;
  }
  if (raw.skill_name !== skillName) {
    errors.push(
      `catalog: evals.json skill_name '${String(raw.skill_name)}' does not match '${String(skillName)}'`,
    );
  }
  if (typeof raw.tier !== "string" || !OUTPUT_TIERS.has(raw.tier)) {
    errors.push(
      "catalog: evals.json tier must be aggressive, moderate, or conservative",
    );
  }
  if (!Array.isArray(raw.evals) || raw.evals.length < 3) {
    errors.push("catalog: evals.json must contain at least 3 output evals");
    return;
  }
  for (const [index, entry] of raw.evals.entries()) {
    if (!isRecord(entry)) {
      errors.push(
        `catalog: evals.json entry ${String(index + 1)} must be an object`,
      );
      continue;
    }
    if (typeof entry.prompt !== "string" || entry.prompt.trim().length === 0) {
      errors.push(
        `catalog: evals.json entry ${String(index + 1)} needs a non-empty prompt`,
      );
    }
    if (
      !Array.isArray(entry.assertions) ||
      entry.assertions.length === 0 ||
      entry.assertions.some(
        (assertion) =>
          typeof assertion !== "string" || assertion.trim().length === 0,
      )
    ) {
      errors.push(
        `catalog: evals.json entry ${String(index + 1)} needs non-empty string assertions`,
      );
    }
  }
  passes.push(
    `catalog: ${String(raw.evals.length)} output eval(s) are structurally valid`,
  );
};

const checkTriggerEvals = (
  skillDir: string,
  passes: string[],
  errors: string[],
  fs: FileSystem,
): void => {
  const path = join(skillDir, "eval-queries.json");
  if (!fs.isFile(path)) {
    errors.push("catalog: missing eval-queries.json");
    return;
  }
  const raw = readJson(path, errors, fs);
  if (!Array.isArray(raw)) {
    if (raw !== undefined)
      errors.push("catalog: eval-queries.json must be an array");
    return;
  }

  let positive = 0;
  let negative = 0;
  const splitCounts = new Map<string, number>();
  const generatedNegativeOwners = new Set<string>();
  let generatedNegativeCount = 0;
  const seen = new Set<string>();
  for (const [index, entry] of raw.entries()) {
    const checked = checkTriggerEntry(entry, index, seen, errors);
    if (checked === undefined) continue;
    if (checked.shouldTrigger) positive += 1;
    else negative += 1;
    if (checked.generatedNegativeOwner !== undefined) {
      generatedNegativeOwners.add(checked.generatedNegativeOwner);
      generatedNegativeCount += 1;
    }
    if (checked.split === undefined) continue;
    const polarity = checked.shouldTrigger ? "positive" : "negative";
    const key = `${polarity}:${checked.split}`;
    splitCounts.set(key, (splitCounts.get(key) ?? 0) + 1);
  }
  checkTriggerMinimums(positive, negative, splitCounts, errors);
  const requiredGeneratedOwners = Math.min(3, generatedNegativeCount);
  if (generatedNegativeOwners.size < requiredGeneratedOwners) {
    errors.push(
      `catalog: generated negative routes need at least ${String(requiredGeneratedOwners)} sibling owners (found ${String(generatedNegativeOwners.size)})`,
    );
  }
  if (errors.every((error) => !error.includes("trigger eval"))) {
    passes.push(
      `catalog: trigger evals cover ${String(positive)} positive and ${String(negative)} negative routes with train/validation splits`,
    );
  }
};

const checkSources = (
  skillDir: string,
  description: unknown,
  passes: string[],
  errors: string[],
  fs: FileSystem,
): void => {
  const path = join(skillDir, "SOURCES.md");
  if (!fs.isFile(path)) {
    errors.push("catalog: missing SOURCES.md");
    return;
  }
  const text = fs.readText(path);
  const hasUrlProvenance = URL_FIELD_RE.test(text) && HTTP_URL_RE.test(text);
  const hasDeclaredProvenance = PROVENANCE_RE.test(text);
  if (!hasUrlProvenance && !hasDeclaredProvenance) {
    errors.push(
      "catalog: SOURCES.md needs an HTTP URL/URLs field or an explicit Provenance field",
    );
  }
  if (!REVIEWED_RE.test(text)) {
    errors.push("catalog: SOURCES.md needs at least one Last reviewed date");
  }
  if (
    typeof description === "string" &&
    VERSIONED_DESCRIPTION_RE.test(description) &&
    !VERSION_FIELD_RE.test(text)
  ) {
    errors.push(
      "catalog: version-pinned skill needs a Version field in SOURCES.md",
    );
  }
  for (const block of text.split(/^##\s+/m).slice(1)) {
    if (!VERSION_FIELD_RE.test(block) || !HTTP_URL_RE.test(block)) continue;
    const hasContentSnapshot = CONTENT_SHA256_FIELD_RE.test(block);
    const hasRepositoryDriftAnchor =
      CHECKOUT_FIELD_RE.test(block) &&
      COMMIT_FIELD_RE.test(block) &&
      WATCH_FIELD_RE.test(block);
    if (!hasContentSnapshot && !hasRepositoryDriftAnchor) {
      const title = block.split(/\r?\n/, 1)[0]?.trim() ?? "unnamed source";
      errors.push(
        `catalog: versioned web source '${title}' needs Content SHA256 or Checkout + Commit + Watch drift fields`,
      );
    }
  }
  if ((hasUrlProvenance || hasDeclaredProvenance) && REVIEWED_RE.test(text)) {
    passes.push("catalog: source provenance and review date are present");
  }
};

const checkCredentialExamples = (
  skillDir: string,
  passes: string[],
  errors: string[],
  fs: FileSystem,
): void => {
  const skill = join(skillDir, "SKILL.md");
  const files: [string, string][] = fs.isFile(skill)
    ? [["SKILL.md", skill]]
    : [];
  const references = join(skillDir, "references");
  if (fs.isDirectory(references)) {
    for (const entry of fs.readDirectory(references).toSorted()) {
      const path = join(references, entry);
      if (entry.endsWith(".md") && fs.isFile(path)) {
        files.push([`references/${entry}`, path]);
      }
    }
  }
  let unsafe = 0;
  for (const [label, path] of files) {
    const text = fs.readText(path);
    for (const [pattern, reason] of UNSAFE_CREDENTIAL_PATTERNS) {
      if (!pattern.test(text)) continue;
      unsafe += 1;
      errors.push(`credentials: ${label} ${reason}`);
    }
  }
  if (unsafe === 0) {
    passes.push("credentials: examples avoid plaintext token anti-patterns");
  }
};

const hasExecutableScripts = (dir: string, fs: FileSystem): boolean => {
  if (!fs.isDirectory(dir)) return false;
  for (const name of fs.readDirectory(dir)) {
    if (name === "__pycache__" || name.startsWith(".")) continue;
    const path = join(dir, name);
    if (fs.isDirectory(path)) {
      if (hasExecutableScripts(path, fs)) return true;
    } else if (fs.isFile(path)) {
      return true;
    }
  }
  return false;
};

const forbiddenImplementationFiles = (
  dir: string,
  fs: FileSystem,
  prefix = "",
): readonly string[] => {
  if (!fs.isDirectory(dir)) return [];
  return fs.readDirectory(dir).flatMap((name) => {
    if (name === "__pycache__" || name.startsWith(".")) return [];
    const path = join(dir, name);
    const label = prefix === "" ? name : `${prefix}/${name}`;
    if (fs.isDirectory(path)) {
      return forbiddenImplementationFiles(path, fs, label);
    }
    return fs.isFile(path) && /\.(?:mjs|ts)$/u.test(name) ? [label] : [];
  });
};

const checkImplementationFormats = (
  skillDir: string,
  errors: string[],
  fs: FileSystem,
): void => {
  const forbidden = forbiddenImplementationFiles(skillDir, fs);
  if (forbidden.length === 0) return;
  errors.push(
    `catalog: skills must not contain TypeScript or MJS implementation files: ${forbidden.join(", ")}`,
  );
};

const checkScriptMetadata = (
  skillDir: string,
  frontmatter: Readonly<Record<string, unknown>>,
  passes: string[],
  errors: string[],
  fs: FileSystem,
): void => {
  if (!hasExecutableScripts(join(skillDir, "scripts"), fs)) return;
  if (
    typeof frontmatter.compatibility !== "string" ||
    frontmatter.compatibility.trim().length === 0
  ) {
    errors.push(
      "catalog: scripted skills need compatibility runtime/network metadata",
    );
  }
  if (
    typeof frontmatter["allowed-tools"] !== "string" ||
    frontmatter["allowed-tools"].trim().length === 0
  ) {
    errors.push(
      "catalog: scripted skills need a non-empty allowed-tools policy",
    );
  }
  if (
    typeof frontmatter.compatibility === "string" &&
    frontmatter.compatibility.trim().length > 0 &&
    typeof frontmatter["allowed-tools"] === "string" &&
    frontmatter["allowed-tools"].trim().length > 0
  ) {
    passes.push(
      "catalog: scripted skill declares compatibility and allowed-tools",
    );
  }
};

export const checkCatalogFiles = (
  skillDir: string,
  frontmatter: Readonly<Record<string, unknown>>,
  fs: FileSystem = nodeFileSystem,
): CatalogFileReport => {
  const passes: string[] = [];
  const errors: string[] = [];
  checkOutputEvals(skillDir, frontmatter.name, passes, errors, fs);
  checkTriggerEvals(skillDir, passes, errors, fs);
  checkSources(skillDir, frontmatter.description, passes, errors, fs);
  checkCredentialExamples(skillDir, passes, errors, fs);
  checkImplementationFormats(skillDir, errors, fs);
  checkScriptMetadata(skillDir, frontmatter, passes, errors, fs);
  return {passes, errors};
};
