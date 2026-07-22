import {existsSync, readdirSync, readFileSync, statSync} from "node:fs";
import {resolve} from "node:path";
import {type z} from "zod";
import {
  LockfileSchema,
  MarketplaceSchema,
  PackageManifestSchema,
  PluginManifestSchema,
  type Lockfile,
  type Marketplace,
  type PackageManifest,
  type PluginManifest,
} from "./release-inputs.js";
import {
  createChecker,
  forbiddenClaims,
  markdownLinkTargets,
  releaseWorkflowFailures,
} from "./validate.js";

export interface ReleaseValidationResult {
  readonly checks: number;
  readonly failures: readonly string[];
  readonly pluginPackages: number;
}

type Check = (condition: boolean, message: string) => void;

const issueDetail = (issue: z.core.$ZodIssue): string => {
  const path =
    issue.path.length > 0 ? issue.path.map(String).join(".") : "root";
  return `${path}: ${issue.message}`;
};

const readJson = <Output>(
  path: string,
  label: string,
  schema: z.ZodType<Output>,
  check: Check,
): Output | undefined => {
  let input: unknown;
  try {
    input = JSON.parse(readFileSync(path, "utf8"));
  } catch (error: unknown) {
    const message = error instanceof Error ? error.message : String(error);
    check(false, `${label} could not be read or parsed: ${message}`);
    return undefined;
  }

  const result = schema.safeParse(input);
  if (!result.success) {
    for (const issue of result.error.issues) {
      check(false, `${label} is invalid: ${issueDetail(issue)}`);
    }
    return undefined;
  }
  return result.data;
};

const childDirectories = (
  path: string,
  label: string,
  check: Check,
): readonly string[] => {
  try {
    return readdirSync(path)
      .map((name) => resolve(path, name))
      .filter((entry) => statSync(entry).isDirectory())
      .toSorted();
  } catch (error: unknown) {
    const message = error instanceof Error ? error.message : String(error);
    check(false, `${label} could not be enumerated: ${message}`);
    return [];
  }
};

const readText = (
  path: string,
  label: string,
  check: Check,
): string | undefined => {
  try {
    return readFileSync(path, "utf8");
  } catch (error: unknown) {
    const message = error instanceof Error ? error.message : String(error);
    check(false, `${label} could not be read: ${message}`);
    return undefined;
  }
};

const marketplaceSource = (
  marketplace: Marketplace,
): ReadonlyMap<string, string | undefined> =>
  new Map(
    marketplace.plugins.map((entry) => [
      entry.name,
      typeof entry.source === "string" ? entry.source : entry.source.path,
    ]),
  );

const marketplaceDescriptions = (
  marketplace: Marketplace,
): ReadonlyMap<string, string | undefined> =>
  new Map(marketplace.plugins.map((entry) => [entry.name, entry.description]));

const dependencyEntries = (
  manifest: PackageManifest,
): readonly (readonly [string, string])[] =>
  Object.entries({
    ...manifest.dependencies,
    ...manifest.devDependencies,
    ...manifest.peerDependencies,
    ...manifest.optionalDependencies,
  });

const skillPaths = (manifest: PluginManifest): readonly string[] =>
  typeof manifest.skills === "string"
    ? [manifest.skills]
    : (manifest.skills ?? []);

const validateSkillPackaging = (
  pluginPackage: string,
  packageLabel: string,
  claudeManifest: PluginManifest,
  codexManifest: PluginManifest,
  check: Check,
): void => {
  const guideNames = readdirSync(pluginPackage)
    .filter((entry) => existsSync(resolve(pluginPackage, entry, "SKILL.md")))
    .toSorted();
  const expectedPaths = guideNames.map((guide) => `./${guide}`);
  const claudePaths = skillPaths(claudeManifest);
  const codexPaths = skillPaths(codexManifest);
  const samePaths = (paths: readonly string[]): boolean =>
    paths.length === expectedPaths.length &&
    paths.toSorted().every((path, index) => path === expectedPaths[index]);

  check(
    Array.isArray(claudeManifest.skills) && claudePaths.length > 0,
    `${packageLabel} Claude manifest exposes skills as a non-empty array`,
  );
  check(
    typeof codexManifest.skills === "string",
    `${packageLabel} Codex manifest exposes one direct skill path`,
  );
  check(
    samePaths(claudePaths),
    `${packageLabel} Claude manifest skill paths match ${expectedPaths.join(", ")}`,
  );
  check(
    samePaths(codexPaths),
    `${packageLabel} Codex manifest skill paths match ${expectedPaths.join(", ")}`,
  );
  for (const path of [...claudePaths, ...codexPaths]) {
    const direct = /^\.\/[a-z0-9]+(?:-[a-z0-9]+)*$/u.test(path);
    check(direct, `${packageLabel} manifest skill path is direct: ${path}`);
    if (direct) {
      check(
        existsSync(resolve(pluginPackage, path, "SKILL.md")),
        `${packageLabel} manifest skill path resolves: ${path}`,
      );
    }
  }
};

const validatePackage = (
  pluginPackage: string,
  repositoryRoot: string,
  releaseVersion: string,
  claudeSources: ReadonlyMap<string, string | undefined>,
  codexSources: ReadonlyMap<string, string | undefined>,
  claudeDescriptions: ReadonlyMap<string, string | undefined>,
  codexDescriptions: ReadonlyMap<string, string | undefined>,
  check: Check,
): void => {
  const relativePath = (path: string): string =>
    path.slice(repositoryRoot.length + 1);
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
    return;
  }

  const packageJson = readJson<PackageManifest>(
    packageJsonPath,
    relativePath(packageJsonPath),
    PackageManifestSchema,
    check,
  );
  const claudeManifest = readJson<PluginManifest>(
    claudeManifestPath,
    relativePath(claudeManifestPath),
    PluginManifestSchema,
    check,
  );
  const codexManifest = readJson<PluginManifest>(
    codexManifestPath,
    relativePath(codexManifestPath),
    PluginManifestSchema,
    check,
  );
  if (
    packageJson === undefined ||
    claudeManifest === undefined ||
    codexManifest === undefined
  ) {
    return;
  }

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
    claudeSources.get(claudeManifest.name) === source,
    `${claudeManifest.name} has the expected Claude marketplace source`,
  );
  check(
    codexSources.get(codexManifest.name) === source,
    `${codexManifest.name} has the expected Codex marketplace source`,
  );

  if (relativePath(pluginPackage).startsWith("packages/skill/")) {
    validateSkillPackaging(
      pluginPackage,
      relativePath(pluginPackage),
      claudeManifest,
      codexManifest,
      check,
    );
  }

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

  for (const [dependency, version] of dependencyEntries(packageJson)) {
    if (/^@xonovex\/(?:command|skill)-/.test(dependency)) {
      check(
        version === releaseVersion,
        `${packageJson.name} dependency ${dependency} matches ${releaseVersion}`,
      );
    }
  }
};

const validateLockfile = (
  repositoryRoot: string,
  releaseVersion: string,
  pluginPackages: readonly string[],
  lockfile: Lockfile,
  check: Check,
): void => {
  for (const pluginPackage of pluginPackages) {
    const key = pluginPackage.slice(repositoryRoot.length + 1);
    check(
      lockfile.packages?.[key]?.version === releaseVersion,
      `package-lock records ${key}@${releaseVersion}`,
    );
  }
};

export const validateRelease = (
  repositoryRoot: string,
): ReleaseValidationResult => {
  const checker = createChecker();
  const check = checker.check;

  const marketplace = readJson<Marketplace>(
    resolve(repositoryRoot, ".claude-plugin/marketplace.json"),
    ".claude-plugin/marketplace.json",
    MarketplaceSchema,
    check,
  );
  const codexMarketplace = readJson<Marketplace>(
    resolve(repositoryRoot, ".agents/plugins/marketplace.json"),
    ".agents/plugins/marketplace.json",
    MarketplaceSchema,
    check,
  );
  if (marketplace === undefined || codexMarketplace === undefined) {
    return {
      checks: checker.checks(),
      failures: checker.failures(),
      pluginPackages: 0,
    };
  }

  const releaseVersion = marketplace.metadata?.version ?? "";
  check(
    /^\d+\.\d+\.\d+$/u.test(releaseVersion),
    "marketplace release must use a semantic version",
  );

  const pluginPackages = [
    ...childDirectories(
      resolve(repositoryRoot, "packages/command"),
      "packages/command",
      check,
    ),
    ...childDirectories(
      resolve(repositoryRoot, "packages/skill"),
      "packages/skill",
      check,
    ),
  ];
  const claudeSources = marketplaceSource(marketplace);
  const codexSources = marketplaceSource(codexMarketplace);
  const claudeDescriptions = marketplaceDescriptions(marketplace);
  const codexDescriptions = marketplaceDescriptions(codexMarketplace);

  for (const pluginPackage of pluginPackages) {
    validatePackage(
      pluginPackage,
      repositoryRoot,
      releaseVersion,
      claudeSources,
      codexSources,
      claudeDescriptions,
      codexDescriptions,
      check,
    );
  }

  check(
    marketplace.plugins.length === pluginPackages.length &&
      claudeSources.size === pluginPackages.length,
    "Claude marketplace contains every command and skill package exactly once",
  );
  check(
    codexMarketplace.plugins.length === pluginPackages.length &&
      codexSources.size === pluginPackages.length,
    "Codex marketplace contains every command and skill package exactly once",
  );

  const lockfile = readJson<Lockfile>(
    resolve(repositoryRoot, "package-lock.json"),
    "package-lock.json",
    LockfileSchema,
    check,
  );
  if (lockfile !== undefined) {
    validateLockfile(
      repositoryRoot,
      releaseVersion,
      pluginPackages,
      lockfile,
      check,
    );
  }

  const rootReadme = readText(
    resolve(repositoryRoot, "README.md"),
    "README.md",
    check,
  );
  if (rootReadme !== undefined) {
    for (const target of markdownLinkTargets(rootReadme)) {
      check(
        existsSync(resolve(repositoryRoot, target)),
        `README.md link resolves: ${target}`,
      );
    }
    for (const forbiddenClaim of forbiddenClaims) {
      check(
        !forbiddenClaim.test(rootReadme),
        `root README rejects ${String(forbiddenClaim)}`,
      );
    }
  }

  const releaseWorkflow = readText(
    resolve(repositoryRoot, ".github/workflows/release.yml"),
    ".github/workflows/release.yml",
    check,
  );
  if (releaseWorkflow !== undefined) {
    for (const failure of releaseWorkflowFailures(releaseWorkflow)) {
      check(false, failure);
    }
  }

  return {
    checks: checker.checks(),
    failures: checker.failures(),
    pluginPackages: pluginPackages.length,
  };
};
