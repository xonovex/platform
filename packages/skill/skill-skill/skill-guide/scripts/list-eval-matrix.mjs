#!/usr/bin/env node
import {execFileSync} from "node:child_process";
import {existsSync, readdirSync, statSync} from "node:fs";
import {join, relative, resolve, sep} from "node:path";

const HELP = `List skill packages for the model-eval workflow.

Usage:
  node scripts/list-eval-matrix.mjs [catalog-root]
  node scripts/list-eval-matrix.mjs [catalog-root] --changed BASE HEAD

Without --changed, emits every skill as a compact JSON array. With --changed,
emits directly changed skills, or the full catalog when shared eval tooling or
workflow configuration changed.`;

const isDirectory = (path) => {
  try {
    return statSync(path).isDirectory();
  } catch {
    return false;
  }
};

const parseArgs = (argv) => {
  if (argv.includes("-h") || argv.includes("--help")) return {help: true};
  const changedIndex = argv.indexOf("--changed");
  const positionals = argv.filter(
    (value, index) =>
      value !== "--changed" &&
      index !== changedIndex + 1 &&
      index !== changedIndex + 2,
  );
  if (positionals.length > 1)
    throw new Error("expected at most one catalog root");
  if (changedIndex !== -1) {
    const base = argv[changedIndex + 1];
    const head = argv[changedIndex + 2];
    if (base === undefined || head === undefined) {
      throw new Error("--changed requires BASE and HEAD revisions");
    }
    return {help: false, root: positionals[0], changed: {base, head}};
  }
  return {help: false, root: positionals[0]};
};

const discoverSkills = (catalogRoot) => {
  const skills = [];
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
    skills.push({
      package: packageName.slice("skill-".length),
      guide: guides[0],
    });
  }
  return skills;
};

const changedFiles = (base, head, workspaceRoot) =>
  execFileSync("git", ["diff", "--name-only", `${base}...${head}`], {
    cwd: workspaceRoot,
    encoding: "utf8",
  })
    .split("\n")
    .map((entry) => entry.trim())
    .filter(Boolean);

const SHARED_PATHS = [
  ".github/workflows/skill-evals.yml",
  ".moon/tasks/tag-skill.yml",
  "packages/script/script-moon-common/",
  "packages/script/script-moon-skill-eval-outputs/",
  "packages/script/script-moon-skill-eval-triggers/",
  "packages/skill/AGENTS.md",
  "packages/skill/skill-skill/skill-guide/scripts/",
];

const selectChangedSkills = (skills, files) => {
  if (
    files.some((file) =>
      SHARED_PATHS.some((path) =>
        path.endsWith("/") ? file.startsWith(path) : file === path,
      ),
    )
  ) {
    return skills;
  }
  const changedPackages = new Set(
    files
      .map((file) => /^packages\/skill\/skill-([^/]+)\//.exec(file)?.[1])
      .filter((value) => value !== undefined),
  );
  return skills.filter((skill) => changedPackages.has(skill.package));
};

const main = () => {
  const args = parseArgs(process.argv.slice(2));
  if (args.help) {
    process.stdout.write(`${HELP}\n`);
    return;
  }
  const workspaceRoot = resolve(".");
  const catalogRoot = resolve(args.root ?? "packages/skill");
  const skills = discoverSkills(catalogRoot);
  const selected =
    args.changed === undefined
      ? skills
      : selectChangedSkills(
          skills,
          changedFiles(args.changed.base, args.changed.head, workspaceRoot),
        );
  process.stdout.write(`${JSON.stringify(selected)}\n`);
};

try {
  main();
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  process.stderr.write(`error: ${message}\n${HELP}\n`);
  process.exitCode = 2;
}
