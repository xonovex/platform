#!/usr/bin/env node
import {existsSync, readdirSync, readFileSync, statSync} from "node:fs";
import {dirname, resolve} from "node:path";
import {fileURLToPath} from "node:url";
import {
  blanketIdBlockFailures,
  createChecker,
  forbiddenClaims,
  legendClasses,
  markdownLinkTargets,
  offLegendClassifications,
  releaseWorkflowFailures,
  tableIds,
  tableShapeFailures,
} from "./validate.js";

interface MarketplaceEntry {
  name: string;
  source: string | {path?: string};
}

interface Marketplace {
  metadata?: {version: string};
  plugins: MarketplaceEntry[];
}

interface PackageManifest {
  name: string;
  version?: string;
  dependencies?: Record<string, string>;
  devDependencies?: Record<string, string>;
  peerDependencies?: Record<string, string>;
  optionalDependencies?: Record<string, string>;
}

interface Lockfile {
  packages?: Record<string, {version?: string}>;
}

interface EvalDocument {
  evals?: {id?: string; assertions?: unknown}[];
}

const scriptDirectory = dirname(fileURLToPath(import.meta.url));
const repositoryRoot = resolve(scriptDirectory, "../../../../..");

const {check, checks, failures} = createChecker();

const read = (path: string): string => readFileSync(path, "utf8");
const readJson = (path: string): unknown => JSON.parse(read(path));
const relativePath = (path: string): string =>
  path.slice(repositoryRoot.length + 1);

const ownerFor = (
  id: string,
  owners: {sources: Set<string>; decisions: Set<string>; controls: Set<string>},
): Set<string> => {
  if (id.startsWith("S-")) return owners.sources;
  if (id.startsWith("D-")) return owners.decisions;
  return owners.controls;
};

const childDirectories = (path: string): string[] =>
  readdirSync(path)
    .map((name) => resolve(path, name))
    .filter((entry) => statSync(entry).isDirectory())
    .toSorted();

const marketplace = readJson(
  resolve(repositoryRoot, ".claude-plugin/marketplace.json"),
) as Marketplace;
const codexMarketplace = readJson(
  resolve(repositoryRoot, ".agents/plugins/marketplace.json"),
) as Marketplace;
const releaseVersion = marketplace.metadata?.version ?? "";

check(
  /^\d+\.\d+\.\d+$/u.test(releaseVersion),
  "marketplace release must use a semantic version",
);

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
    typeof entry.source === "string" ? entry.source : entry.source.path,
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

  const packageJson = readJson(packageJsonPath) as PackageManifest;
  const claudeManifest = readJson(claudeManifestPath) as PackageManifest;
  const codexManifest = readJson(codexManifestPath) as PackageManifest;
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

const lockfile = readJson(
  resolve(repositoryRoot, "package-lock.json"),
) as Lockfile;
for (const pluginPackage of pluginPackages) {
  const key = relativePath(pluginPackage);
  check(
    lockfile.packages?.[key]?.version === releaseVersion,
    `package-lock records ${key}@${releaseVersion}`,
  );
}

const rootReadmePath = resolve(repositoryRoot, "README.md");
const rootReadme = read(rootReadmePath);
for (const target of markdownLinkTargets(rootReadme)) {
  check(
    existsSync(resolve(repositoryRoot, target)),
    `README.md link resolves: ${target}`,
  );
}

const releaseWorkflow = read(
  resolve(repositoryRoot, ".github/workflows/release.yml"),
);
for (const failure of releaseWorkflowFailures(releaseWorkflow)) {
  check(false, failure);
}
for (const forbiddenClaim of forbiddenClaims) {
  check(
    !forbiddenClaim.test(rootReadme),
    `root README rejects ${String(forbiddenClaim)}`,
  );
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

const sourceIds = tableIds(sourceRegistry, "S-");
const decisionIds = tableIds(decisionMatrix, "D-");
const controlIds = tableIds(controlCrosswalk, "C-");
check(
  sourceIds.size === 94,
  `source registry resolves 94 IDs, found ${String(sourceIds.size)}`,
);
check(
  decisionIds.size === 39,
  `decision matrix resolves 39 IDs, found ${String(decisionIds.size)}`,
);
check(
  controlIds.size === 45,
  `control crosswalk resolves 45 IDs, found ${String(controlIds.size)}`,
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
  const owner = ownerFor(reference, {
    sources: sourceIds,
    decisions: decisionIds,
    controls: controlIds,
  });
  check(owner.has(reference), `traceability reference resolves: ${reference}`);
}

const subplanFiles = readdirSync(planRoot)
  .filter((name) => /^subplan-\d{2}-.*\.md$/.test(name))
  .toSorted();
check(
  subplanFiles.length === 11,
  `plan set contains 11 subplans, found ${String(subplanFiles.length)}`,
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
    const owner = ownerFor(id, {
      sources: sourceIds,
      decisions: decisionIds,
      controls: controlIds,
    });
    check(
      owner.has(id),
      `${subplanFile} frontmatter reference resolves: ${id}`,
    );
  }
}
check(
  taskCount === 150,
  `subplans contain 150 numbered tasks, found ${String(taskCount)}`,
);

const taskRows =
  subplanTraceability.match(/^\| subplan-\d{2}-.*\|\s*\d+\s*\|/gm) ?? [];
check(
  taskRows.length === taskCount,
  `task traceability has ${String(taskCount)} rows, found ${String(taskRows.length)}`,
);
const platformRows =
  platformMatrix.match(/^\| (?!---|Platform)[^|]+\|/gm) ?? [];
check(
  platformRows.length === 14,
  `platform matrix has 14 entries, found ${String(platformRows.length)}`,
);

check(
  /current consolidated text/i.test(sourceRegistry) &&
    /qualified (?:legal |privacy\/legal )?review/i.test(sourceRegistry),
  "source registry preserves current-text and qualified-review flags",
);
check(
  /licensed (?:full )?text/i.test(sourceRegistry) &&
    validationPolicy.includes("licensed-text-required"),
  "source registry and policy preserve licensed-standard flags",
);

const shapeTargets: [string, string][] = [
  [sourceRegistry, "source-registry.md"],
  [decisionMatrix, "decision-source-matrix.md"],
  [controlCrosswalk, "control-crosswalk.md"],
  [platformMatrix, "platform-capability-matrix.md"],
  [subplanTraceability, "subplan-traceability.md"],
];
for (const [content, label] of shapeTargets) {
  const shapeFailures = tableShapeFailures(content, label);
  check(
    shapeFailures.length === 0,
    `table shapes valid in ${label}: ${shapeFailures.join("; ")}`,
  );
}

const legend = legendClasses(decisionMatrix);
check(
  legend.size === 4,
  `decision legend defines 4 classes, found ${String(legend.size)}`,
);
const offLegend = offLegendClassifications(decisionMatrix, legend);
check(
  offLegend.length === 0,
  `decision classifications stay in the legend: ${offLegend.join("; ")}`,
);

const blanketBlocks = blanketIdBlockFailures(subplanTraceability);
check(
  blanketBlocks.length === 0,
  `no subplan repeats one identical ID block on every task: ${blanketBlocks.join("; ")}`,
);

const parentPlan = read(
  resolve(repositoryRoot, "plans/composable-workflow-phases.md"),
);
check(
  /eleven child plans include one deliberate exception/i.test(parentPlan),
  "parent plan documents the deliberate eleven-subplan exception",
);

const evalFiles = [
  "packages/skill/skill-claude-code/code-harness-guide/evals.json",
  "packages/skill/skill-codex/codex-guide/evals.json",
  "packages/skill/skill-kiro/kiro-guide/evals.json",
  "packages/skill/skill-copilot/copilot-guide/evals.json",
  "packages/skill/skill-pi/pi-guide/evals.json",
  "packages/skill/skill-opencode/opencode-guide/evals.json",
  "packages/skill/skill-azure-devops/azure-devops-guide/evals.json",
  "packages/skill/skill-bitbucket/bitbucket-guide/evals.json",
  "packages/skill/skill-bitrise/bitrise-guide/evals.json",
  "packages/skill/skill-aws/aws-guide/evals.json",
  "packages/skill/skill-datadog/datadog-guide/evals.json",
];
for (const evalFile of evalFiles) {
  const evalPath = resolve(repositoryRoot, evalFile);
  check(existsSync(evalPath), `${evalFile} exists`);
  if (!existsSync(evalPath)) continue;
  const evalDocument = readJson(evalPath) as EvalDocument;
  check(
    (evalDocument.evals?.length ?? 0) >= 3,
    `${evalFile} has at least three output evals`,
  );
  for (const evalCase of evalDocument.evals ?? []) {
    check(
      Array.isArray(evalCase.assertions) && evalCase.assertions.length > 0,
      `${evalFile} eval ${evalCase.id ?? "unknown"} has observable assertions`,
    );
  }
}

const conformanceAssets = [
  "packages/skill/skill-agent-governance/agent-governance-guide/scripts/workflow-runtime.test.ts",
  "packages/skill/skill-agent-governance/agent-governance-guide/scripts/workflow-command-runtime.test.ts",
  "packages/skill/skill-agent-governance/agent-governance-guide/scripts/workflow-trigger-adapters.test.ts",
  "packages/skill/skill-agent-governance/agent-governance-guide/scripts/workflow-maturity.test.ts",
  "packages/skill/skill-workflow/workflow-guide/assets/conformance-fixtures.json",
  "packages/skill/skill-workflow/workflow-guide/assets/development-assurance-fixtures.json",
  "packages/skill/skill-workflow/workflow-guide/assets/operational-lifecycle-fixtures.json",
].map((path) => resolve(repositoryRoot, path));
const conformanceText = conformanceAssets.map(read).join("\n").toLowerCase();
for (const coverageTerm of [
  "concurrency",
  "idempotency",
  "capability",
  "executor",
  "trigger",
  "observe",
  "enforce",
  "evidence",
  "authority",
  "rollback",
  "policy",
  "lifecycle",
  "provider",
]) {
  check(
    conformanceText.includes(coverageTerm),
    `conformance fixtures cover ${coverageTerm}`,
  );
}

const diagramDirectory = resolve(
  repositoryRoot,
  "packages/diagram/diagram-agent-workflow",
);
const diagramSources = ["workflow-diagram.dot", "target-architecture.dot"].map(
  (name) => read(resolve(diagramDirectory, name)),
);
const combinedDiagrams = diagramSources.join("\n").toLowerCase();
for (const diagramTerm of [
  "neutral workflow kernel",
  "trigger adapter",
  "executor port",
  "control port",
  "evidence port",
  "trusted registry",
  "independent host",
  "derived maturity",
]) {
  check(
    combinedDiagrams.includes(diagramTerm),
    `diagrams include ${diagramTerm}`,
  );
}
for (const image of ["workflow-diagram.png", "target-architecture.png"]) {
  const imagePath = resolve(diagramDirectory, image);
  check(
    existsSync(imagePath) && statSync(imagePath).size > 0,
    `${image} exists and is non-empty`,
  );
}

if (failures().length > 0) {
  console.error(
    `Release validation failed: ${String(failures().length)}/${String(checks())} checks`,
  );
  for (const failure of failures()) console.error(`- ${failure}`);
  process.exit(1);
}

console.log(
  `Release validation passed: ${String(checks())} checks, ${String(pluginPackages.length)} lockstep packages, ${String(sourceIds.size)} sources, ${String(decisionIds.size)} decisions, ${String(controlIds.size)} controls, ${String(taskCount)} mapped tasks`,
);
