#!/usr/bin/env node
import {existsSync, readdirSync, readFileSync} from "node:fs";
import {dirname, extname, resolve} from "node:path";
import {fileURLToPath} from "node:url";

const scriptDirectory = dirname(fileURLToPath(import.meta.url));
const packageDirectory = resolve(scriptDirectory, "..");

const failures = [];
let checks = 0;

const check = (condition, message) => {
  checks += 1;
  if (!condition) failures.push(message);
};

const read = (path) => readFileSync(path, "utf8");
const readJson = (path) => JSON.parse(read(path));
const relativePath = (path) => path.slice(packageDirectory.length + 1);

const markdownFiles = (path) => {
  const entries = readdirSync(path, {withFileTypes: true});
  return entries.flatMap((entry) => {
    const entryPath = resolve(path, entry.name);
    if (entry.isDirectory()) return markdownFiles(entryPath);
    return extname(entry.name) === ".md" ? [entryPath] : [];
  });
};

const workflowPackage = readJson(resolve(packageDirectory, "package.json"));
const workflowClaudeManifest = readJson(
  resolve(packageDirectory, ".claude-plugin/plugin.json"),
);
const workflowCodexManifest = readJson(
  resolve(packageDirectory, ".codex-plugin/plugin.json"),
);
check(
  JSON.stringify(workflowClaudeManifest.dependencies) ===
    JSON.stringify(workflowCodexManifest.dependencies),
  "workflow plugin dependencies match across harness manifests",
);
for (const dependency of workflowClaudeManifest.dependencies) {
  const packageName = dependency.replace("xonovex-skill-", "@xonovex/skill-");
  check(
    workflowPackage.dependencies?.[packageName] === workflowPackage.version,
    `workflow package declares ${packageName}@${workflowPackage.version}`,
  );
}

const documentationFiles = [
  resolve(packageDirectory, "README.md"),
  resolve(packageDirectory, "MIGRATION.md"),
  ...markdownFiles(resolve(packageDirectory, "docs")),
];

for (const documentationFile of documentationFiles) {
  const content = read(documentationFile);
  const linkPattern = /\[[^\]]+\]\(([^)]+)\)/g;
  for (const match of content.matchAll(linkPattern)) {
    const target = match[1].split("#", 1)[0];
    if (!target || /^(?:https?:|mailto:)/.test(target)) continue;
    const resolvedTarget = resolve(dirname(documentationFile), target);
    check(
      existsSync(resolvedTarget),
      `${relativePath(documentationFile)} link resolves: ${target}`,
    );
  }
}

const migration = read(resolve(packageDirectory, "MIGRATION.md"));
for (const requiredMigrationTerm of [
  "story-refine",
  "acceptance-formalize",
  "plan-decide",
  "acceptance-decide",
  "Rollback",
  "5.0.0",
  "6.0.0",
]) {
  check(
    migration.includes(requiredMigrationTerm),
    `migration guide covers ${requiredMigrationTerm}`,
  );
}

const validationGuide = read(
  resolve(packageDirectory, "docs/validation-and-traceability.md"),
);
for (const matrixTerm of [
  "Workflow contracts",
  "Governance contracts",
  "Harness adapters",
  "Module execution",
  "Onboarding",
  "External enforcement",
  "Enterprise providers",
  "Reliability and security",
  "Documentation and release",
]) {
  check(
    validationGuide.includes(matrixTerm),
    `validation matrix covers ${matrixTerm}`,
  );
}

const publishedMarkdown = documentationFiles.map(read).join("\n");
const forbiddenClaims = [
  /all harnesses (?:have|support|provide)/i,
  /skills? (?:are|is|provide|provides) enforcement/i,
  /install(?:ing|ed)? (?:a )?skills? (?:enforces|proves)/i,
  /workflow ya?ml (?:is|are) required/i,
  /(?:provides?|ensures?|establishes?|achieves?) automatic compliance/i,
  /silently launch(?:es|ing)? (?:a |an )?(?:child )?agent/i,
];
for (const forbiddenClaim of forbiddenClaims) {
  check(
    !forbiddenClaim.test(publishedMarkdown),
    `published docs reject ${forbiddenClaim}`,
  );
}

if (failures.length > 0) {
  console.error(
    `Documentation validation failed: ${failures.length}/${checks} checks`,
  );
  for (const failure of failures) console.error(`- ${failure}`);
  process.exit(1);
}

console.log(`Documentation validation passed: ${checks} package checks`);
