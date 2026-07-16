#!/usr/bin/env node
import {existsSync, readdirSync, readFileSync, statSync} from "node:fs";
import {dirname, extname, resolve} from "node:path";
import {fileURLToPath} from "node:url";

const scriptDirectory = dirname(fileURLToPath(import.meta.url));
const packageDirectory = resolve(scriptDirectory, "..");
const repositoryRoot = resolve(packageDirectory, "../../..");

const failures = [];
let checks = 0;

const check = (condition, message) => {
  checks += 1;
  if (!condition) failures.push(message);
};

const read = (path) => readFileSync(path, "utf8");
const readJson = (path) => JSON.parse(read(path));
const relativePath = (path) => path.slice(repositoryRoot.length + 1);

const childDirectories = (path) =>
  readdirSync(path)
    .map((name) => resolve(path, name))
    .filter((entry) => statSync(entry).isDirectory())
    .sort();

const markdownFiles = (path) => {
  const entries = readdirSync(path, {withFileTypes: true});
  return entries.flatMap((entry) => {
    const entryPath = resolve(path, entry.name);
    if (entry.isDirectory()) return markdownFiles(entryPath);
    return extname(entry.name) === ".md" ? [entryPath] : [];
  });
};

const marketplacePath = resolve(
  repositoryRoot,
  ".claude-plugin/marketplace.json",
);
const codexMarketplacePath = resolve(
  repositoryRoot,
  ".agents/plugins/marketplace.json",
);
const marketplace = readJson(marketplacePath);
const codexMarketplace = readJson(codexMarketplacePath);
const releaseVersion = marketplace.metadata.version;

check(releaseVersion === "6.0.0", "marketplace release must be 6.0.0");

const pluginPackages = [
  ...childDirectories(resolve(repositoryRoot, "packages/command")),
  ...childDirectories(resolve(repositoryRoot, "packages/skill")),
];

const claudeEntries = new Map(
  marketplace.plugins.map((entry) => [entry.name, entry.source]),
);
const codexEntries = new Map(
  codexMarketplace.plugins.map((entry) => [
    entry.name,
    entry.source?.path ?? entry.source,
  ]),
);

for (const pluginPackage of pluginPackages) {
  const packageJsonPath = resolve(pluginPackage, "package.json");
  const claudeManifestPath = resolve(
    pluginPackage,
    ".claude-plugin/plugin.json",
  );
  const codexManifestPath = resolve(pluginPackage, ".codex-plugin/plugin.json");

  check(
    existsSync(packageJsonPath),
    `${relativePath(pluginPackage)} has package.json`,
  );
  check(
    existsSync(claudeManifestPath),
    `${relativePath(pluginPackage)} has a Claude plugin manifest`,
  );
  check(
    existsSync(codexManifestPath),
    `${relativePath(pluginPackage)} has a Codex plugin manifest`,
  );
  if (
    !existsSync(packageJsonPath) ||
    !existsSync(claudeManifestPath) ||
    !existsSync(codexManifestPath)
  ) {
    continue;
  }

  const packageJson = readJson(packageJsonPath);
  const claudeManifest = readJson(claudeManifestPath);
  const codexManifest = readJson(codexManifestPath);
  const source = `./${relativePath(pluginPackage)}`;

  check(
    packageJson.version === releaseVersion,
    `${packageJson.name} package version matches ${releaseVersion}`,
  );
  check(
    claudeManifest.version === releaseVersion,
    `${claudeManifest.name} Claude manifest version matches ${releaseVersion}`,
  );
  check(
    codexManifest.version === releaseVersion,
    `${codexManifest.name} Codex manifest version matches ${releaseVersion}`,
  );
  check(
    claudeManifest.name === codexManifest.name,
    `${packageJson.name} manifest names match`,
  );
  check(
    claudeEntries.get(claudeManifest.name) === source,
    `${claudeManifest.name} has the expected Claude marketplace source`,
  );
  check(
    codexEntries.get(codexManifest.name) === source,
    `${codexManifest.name} has the expected Codex marketplace source`,
  );

  for (const [dependency, version] of Object.entries({
    ...packageJson.dependencies,
    ...packageJson.devDependencies,
    ...packageJson.peerDependencies,
    ...packageJson.optionalDependencies,
  })) {
    if (/^@xonovex\/(?:command|skill)-/.test(dependency)) {
      check(
        version === releaseVersion,
        `${packageJson.name} dependency ${dependency} matches ${releaseVersion}`,
      );
    }
  }
}

check(
  claudeEntries.size === pluginPackages.length,
  "Claude marketplace contains every command and skill package exactly once",
);
check(
  codexEntries.size === pluginPackages.length,
  "Codex marketplace contains every command and skill package exactly once",
);

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
    workflowPackage.dependencies?.[packageName] === releaseVersion,
    `workflow package declares ${packageName}@${releaseVersion}`,
  );
}

const lockfile = readJson(resolve(repositoryRoot, "package-lock.json"));
for (const pluginPackage of pluginPackages) {
  const key = relativePath(pluginPackage);
  check(
    lockfile.packages?.[key]?.version === releaseVersion,
    `package-lock records ${key}@${releaseVersion}`,
  );
}

const documentationFiles = [
  resolve(repositoryRoot, "README.md"),
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

const planRoot = resolve(repositoryRoot, "plans/composable-workflow-phases");
const traceabilityRoot = resolve(planRoot, "traceability");
const sourceRegistry = read(resolve(traceabilityRoot, "source-registry.md"));
const decisionMatrix = read(
  resolve(traceabilityRoot, "decision-source-matrix.md"),
);
const controlCrosswalk = read(
  resolve(traceabilityRoot, "control-crosswalk.md"),
);
const platformMatrix = read(
  resolve(traceabilityRoot, "platform-capability-matrix.md"),
);
const subplanTraceability = read(
  resolve(traceabilityRoot, "subplan-traceability.md"),
);
const validationPolicy = read(
  resolve(traceabilityRoot, "validation-policy.md"),
);

const tableIds = (content, prefix) =>
  new Set(
    [
      ...content.matchAll(
        new RegExp(`^\\|\\s*(${prefix}[^|\\s]+)\\s*\\|`, "gm"),
      ),
    ].map((match) => match[1]),
  );

const sourceIds = tableIds(sourceRegistry, "S-");
const decisionIds = tableIds(decisionMatrix, "D-");
const controlIds = tableIds(controlCrosswalk, "C-");
check(
  sourceIds.size === 94,
  `source registry resolves 94 IDs, found ${sourceIds.size}`,
);
check(
  decisionIds.size === 39,
  `decision matrix resolves 39 IDs, found ${decisionIds.size}`,
);
check(
  controlIds.size === 45,
  `control crosswalk resolves 45 IDs, found ${controlIds.size}`,
);

const allTraceability = [
  decisionMatrix,
  controlCrosswalk,
  platformMatrix,
  subplanTraceability,
].join("\n");
for (const reference of allTraceability.match(
  /\b(?:S-[A-Z0-9-]+|D-\d{3}|C-\d{3})\b/g,
) ?? []) {
  const owner = reference.startsWith("S-")
    ? sourceIds
    : reference.startsWith("D-")
      ? decisionIds
      : controlIds;
  check(owner.has(reference), `traceability reference resolves: ${reference}`);
}

const subplanFiles = readdirSync(planRoot)
  .filter((name) => /^subplan-\d{2}-.*\.md$/.test(name))
  .sort();
check(
  subplanFiles.length === 11,
  `plan set contains 11 subplans, found ${subplanFiles.length}`,
);

let taskCount = 0;
for (const subplanFile of subplanFiles) {
  const content = read(resolve(planRoot, subplanFile));
  const tasksSection = content
    .slice(content.indexOf("## Tasks"), content.indexOf("## Traceability"))
    .trim();
  const numberedTasks = tasksSection.match(/^\d+\.\s+/gm) ?? [];
  taskCount += numberedTasks.length;
  for (const reference of content.match(
    /^\s*-\s+(?:S-[A-Z0-9-]+|D-\d{3}|C-\d{3})\s*$/gm,
  ) ?? []) {
    const id = reference.replace(/^\s*-\s+/, "").trim();
    const owner = id.startsWith("S-")
      ? sourceIds
      : id.startsWith("D-")
        ? decisionIds
        : controlIds;
    check(
      owner.has(id),
      `${subplanFile} frontmatter reference resolves: ${id}`,
    );
  }
}
check(
  taskCount === 150,
  `subplans contain 150 numbered tasks, found ${taskCount}`,
);

const taskRows =
  subplanTraceability.match(/^\| subplan-\d{2}-.*\|\s*\d+\s*\|/gm) ?? [];
check(
  taskRows.length === taskCount,
  `task traceability has ${taskCount} rows, found ${taskRows.length}`,
);
const platformRows =
  platformMatrix.match(/^\| (?!---|Platform)[^|]+\|/gm) ?? [];
check(
  platformRows.length === 14,
  `platform matrix has 14 entries, found ${platformRows.length}`,
);

check(
  /current consolidated text/i.test(sourceRegistry) &&
    /qualified (?:legal |privacy\/legal )?review/i.test(sourceRegistry),
  "source registry preserves current-text and qualified-review flags",
);
check(
  /licensed (?:full )?text/i.test(sourceRegistry) &&
    /licensed-text-required/.test(validationPolicy),
  "source registry and policy preserve licensed-standard flags",
);

const parentPlan = read(
  resolve(repositoryRoot, "plans/composable-workflow-phases.md"),
);
check(
  /eleven child plans include one deliberate exception/i.test(parentPlan),
  "parent plan documents the deliberate eleven-subplan exception",
);

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

const evalFiles = [
  "packages/skill/skill-claude-code/code-harness-guide/evals.json",
  "packages/skill/skill-codex/codex-guide/evals.json",
  "packages/skill/skill-kiro/kiro-guide/evals.json",
  "packages/skill/skill-copilot/copilot-guide/evals.json",
  "packages/skill/skill-pi/pi-guide/evals.json",
  "packages/skill/skill-opencode/opencode-guide/evals.json",
  "packages/skill/skill-azure-devops/skills/azure-devops-guide/evals.json",
  "packages/skill/skill-bitbucket/skills/bitbucket-guide/evals.json",
  "packages/skill/skill-bitrise/skills/bitrise-guide/evals.json",
  "packages/skill/skill-aws/skills/aws-guide/evals.json",
  "packages/skill/skill-datadog/skills/datadog-guide/evals.json",
];
for (const evalFile of evalFiles) {
  const evalPath = resolve(repositoryRoot, evalFile);
  check(existsSync(evalPath), `${evalFile} exists`);
  if (!existsSync(evalPath)) continue;
  const evalDocument = readJson(evalPath);
  check(
    evalDocument.evals?.length >= 3,
    `${evalFile} has at least three output evals`,
  );
  for (const evalCase of evalDocument.evals ?? []) {
    check(
      Array.isArray(evalCase.assertions) && evalCase.assertions.length > 0,
      `${evalFile} eval ${evalCase.id} has observable assertions`,
    );
  }
}

const conformanceAssets = [
  "packages/skill/skill-agent-governance/agent-governance-guide/assets/conformance-fixtures.json",
  "packages/skill/skill-agent-governance/agent-governance-guide/assets/harness-conformance-fixtures.json",
  "packages/skill/skill-agent-governance/agent-governance-guide/assets/external-enforcement-fixtures.json",
  "packages/skill/skill-agent-governance/agent-governance-guide/assets/enterprise-platform-fixtures.json",
  "packages/skill/skill-agent-governance/agent-governance-guide/assets/walking-skeleton-fixtures.json",
  "packages/skill/skill-agent-governance/agent-governance-guide/assets/governance-operations-fixtures.json",
  "packages/skill/skill-workflow/workflow-guide/assets/conformance-fixtures.json",
  "packages/skill/skill-workflow/workflow-guide/assets/development-assurance-fixtures.json",
  "packages/skill/skill-workflow/workflow-guide/assets/operational-lifecycle-fixtures.json",
].map((path) => resolve(repositoryRoot, path));
const conformanceText = conformanceAssets.map(read).join("\n").toLowerCase();
for (const coverageTerm of [
  "concurrency",
  "idempotency",
  "recursion",
  "authority",
  "rollback",
  "drift",
  "telemetry",
  "onboarding",
  "policy",
  "lifecycle",
  "provider",
]) {
  check(
    conformanceText.includes(coverageTerm),
    `conformance fixtures cover ${coverageTerm}`,
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

const diagramDirectory = resolve(
  repositoryRoot,
  "packages/diagram/diagram-agent-workflow",
);
const diagramSources = [
  "workflow-diagram.dot",
  "target-architecture.dot",
  "maturity-ladder.dot",
].map((name) => read(resolve(diagramDirectory, name)));
const combinedDiagrams = diagramSources.join("\n").toLowerCase();
for (const diagramTerm of [
  "workflow plane",
  "governance plane",
  "semantic event intents",
  "native adapters",
  "external enforcement",
  "provider-native evidence",
  "onboarding",
  "feedback",
]) {
  check(
    combinedDiagrams.includes(diagramTerm),
    `diagrams include ${diagramTerm}`,
  );
}
for (const image of [
  "workflow-diagram.png",
  "target-architecture.png",
  "maturity-ladder.png",
]) {
  const imagePath = resolve(diagramDirectory, image);
  check(
    existsSync(imagePath) && statSync(imagePath).size > 0,
    `${image} exists and is non-empty`,
  );
}

if (failures.length > 0) {
  console.error(
    `Documentation validation failed: ${failures.length}/${checks} checks`,
  );
  for (const failure of failures) console.error(`- ${failure}`);
  process.exit(1);
}

console.log(
  `Documentation validation passed: ${checks} checks, ${pluginPackages.length} lockstep packages, ${sourceIds.size} sources, ${decisionIds.size} decisions, ${controlIds.size} controls, ${taskCount} mapped tasks`,
);
