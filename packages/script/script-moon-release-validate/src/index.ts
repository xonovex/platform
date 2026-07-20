#!/usr/bin/env node
import {existsSync, readdirSync, readFileSync, statSync} from "node:fs";
import {dirname, resolve} from "node:path";
import {fileURLToPath} from "node:url";
import {
  createChecker,
  forbiddenClaims,
  markdownLinkTargets,
  releaseWorkflowFailures,
} from "./validate.js";

interface MarketplaceEntry {
  name: string;
  source: string | {path?: string};
  description?: string;
}

interface Marketplace {
  metadata?: {version: string};
  plugins: MarketplaceEntry[];
}

interface PackageManifest {
  name: string;
  version?: string;
  description?: string;
  dependencies?: Record<string, string>;
  devDependencies?: Record<string, string>;
  peerDependencies?: Record<string, string>;
  optionalDependencies?: Record<string, string>;
}

interface Lockfile {
  packages?: Record<string, {version?: string}>;
}

const scriptDirectory = dirname(fileURLToPath(import.meta.url));
const repositoryRoot = resolve(scriptDirectory, "../../../../..");

const {check, checks, failures} = createChecker();

const read = (path: string): string => readFileSync(path, "utf8");
const readJson = (path: string): unknown => JSON.parse(read(path));
const relativePath = (path: string): string =>
  path.slice(repositoryRoot.length + 1);

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
const claudeDescriptions = new Map(
  marketplace.plugins.map((entry) => [entry.name, entry.description]),
);
const codexDescriptions = new Map(
  codexMarketplace.plugins.map((entry) => [entry.name, entry.description]),
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
  if (packageJson.description !== undefined) {
    check(
      claudeManifest.description === packageJson.description,
      `${packageJson.name} package and Claude manifest descriptions match`,
    );
    check(
      codexManifest.description === packageJson.description,
      `${packageJson.name} package and Codex manifest descriptions match`,
    );
    check(
      claudeDescriptions.get(claudeManifest.name) === packageJson.description,
      `${packageJson.name} package and Claude marketplace descriptions match`,
    );
    check(
      codexDescriptions.get(codexManifest.name) === packageJson.description,
      `${packageJson.name} package and Codex marketplace descriptions match`,
    );
  }

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

const rootReadme = read(resolve(repositoryRoot, "README.md"));
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

if (failures().length > 0) {
  console.error(
    `Release validation failed: ${String(failures().length)}/${String(checks())} checks`,
  );
  for (const failure of failures()) console.error(`- ${failure}`);
  process.exit(1);
}

console.log(
  `Release validation passed: ${String(checks())} checks, ${String(pluginPackages.length)} lockstep packages`,
);
