import {execFileSync} from "node:child_process";
import {existsSync, readdirSync, statSync} from "node:fs";
import {join} from "node:path";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";

export interface SkillSelection {
  readonly package: string;
  readonly guide: string;
}

type JsonRecord = Readonly<Record<string, unknown>>;

const isDirectory = (path: string): boolean => {
  try {
    return statSync(path).isDirectory();
  } catch {
    return false;
  }
};

const asRecord = (value: unknown, description: string): JsonRecord => {
  if (typeof value !== "object" || value === null || Array.isArray(value)) {
    throw new Error(`${description} must be a JSON object`);
  }
  return value as JsonRecord;
};

const parseJson = (content: string, description: string): unknown => {
  try {
    return JSON.parse(content) as unknown;
  } catch (error) {
    const detail = error instanceof Error ? error.message : String(error);
    throw new Error(`${description} is not valid JSON: ${detail}`);
  }
};

export const discoverSkills = (catalogRoot: string): SkillSelection[] => {
  const skills: SkillSelection[] = [];
  for (const packageName of readdirSync(catalogRoot).toSorted()) {
    if (!packageName.startsWith("skill-")) continue;
    const packageDirectory = join(catalogRoot, packageName);
    if (!isDirectory(packageDirectory)) continue;
    const guides = readdirSync(packageDirectory)
      .filter((entry) => existsSync(join(packageDirectory, entry, "SKILL.md")))
      .toSorted();
    if (guides.length !== 1) {
      throw new Error(
        `${packageDirectory} must contain exactly one guide, found ${String(guides.length)}`,
      );
    }
    const guide = guides[0];
    if (guide !== undefined) {
      skills.push({
        package: packageName.slice("skill-".length),
        guide,
      });
    }
  }
  return skills;
};

export const SHARED_SKILL_EVAL_PATHS = [
  ".github/workflows/skill-evals.yml",
  ".moon/tasks/tag-skill.yml",
  "packages/script/script-moon-ci-workflows/",
  "packages/script/script-moon-common/",
  "packages/script/script-moon-skill-eval-outputs/",
  "packages/script/script-moon-skill-eval-triggers/",
  "packages/skill/AGENTS.md",
  "packages/skill/skill-skill/skill-guide/scripts/",
] as const;

const isSharedSkillEvalPath = (file: string): boolean =>
  SHARED_SKILL_EVAL_PATHS.some((path) =>
    path.endsWith("/") ? file.startsWith(path) : file === path,
  );

export const selectChangedSkills = (
  skills: readonly SkillSelection[],
  files: readonly string[],
): SkillSelection[] => {
  if (files.some(isSharedSkillEvalPath)) return [...skills];

  const changedPackages = new Set(
    files.flatMap((file) => {
      const packageName = /^packages\/skill\/skill-([^/]+)\//u.exec(file)?.[1];
      return packageName === undefined ? [] : [packageName];
    }),
  );
  return skills.filter((skill) => changedPackages.has(skill.package));
};

export const changedFiles = (
  base: string,
  head: string,
  workspaceRoot: string,
): string[] =>
  execFileSync(
    resolveExecutable("git"),
    ["diff", "--name-only", `${base}...${head}`],
    {
      cwd: workspaceRoot,
      encoding: "utf8",
    },
  )
    .split("\n")
    .map((entry) => entry.trim())
    .filter(Boolean);

const chunks = <T>(items: readonly T[], size: number): T[][] => {
  if (!Number.isInteger(size) || size < 1) {
    throw new Error("batch size must be a positive integer");
  }
  const result: T[][] = [];
  for (let offset = 0; offset < items.length; offset += size) {
    result.push(items.slice(offset, offset + size));
  }
  return result;
};

export const triggerEvalBatches = (
  content: string,
  batchSize = 8,
): readonly JsonRecord[][] => {
  const parsed = parseJson(content, "trigger eval source");
  if (!Array.isArray(parsed)) {
    throw new TypeError("trigger eval source must be a JSON array");
  }
  const validationQueries = parsed
    .map((entry) => asRecord(entry, "trigger eval entry"))
    .filter((entry) => entry.split === "validation");
  return chunks(validationQueries, batchSize);
};

export const outputEvalBatches = (
  content: string,
  batchSize = 6,
): readonly JsonRecord[] => {
  const document = asRecord(
    parseJson(content, "output eval source"),
    "output eval source",
  );
  if (!Array.isArray(document.evals)) {
    throw new TypeError("output eval source must contain an evals array");
  }
  const evals = document.evals.map((entry) =>
    asRecord(entry, "output eval entry"),
  );
  return chunks(evals, batchSize).map((batch) => ({...document, evals: batch}));
};

export const assertSkillSegment = (value: string, label: string): void => {
  if (!/^[a-z0-9]+(?:-[a-z0-9]+)*$/u.test(value)) {
    throw new Error(`${label} must be a lowercase kebab-case name`);
  }
};
