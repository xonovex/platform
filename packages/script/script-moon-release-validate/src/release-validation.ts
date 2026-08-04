import {execFileSync} from "node:child_process";
import {existsSync, readdirSync, readFileSync, statSync} from "node:fs";
import {basename, resolve} from "node:path";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";
import {type z} from "zod";
import {coverageFloorFailures} from "./coverage-floors.js";
import {workspaceHasherFailures} from "./hasher-ignore.js";
import {
  instructionDocFailures,
  type InstructionDirectory,
} from "./instruction-docs.js";
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
import {taskInheritanceFailures} from "./task-inheritance.js";
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

// gitIgnoredPaths lists what git excludes from the working tree, collapsing a
// fully ignored directory to a single entry. Returns undefined when git cannot
// answer, so a checkout without a git directory reports one clear failure
// rather than a missing-coverage failure per pattern.
const gitIgnoredPaths = (
  repositoryRoot: string,
): readonly string[] | undefined => {
  try {
    return execFileSync(
      resolveExecutable("git"),
      [
        "ls-files",
        "--others",
        "--ignored",
        "--exclude-standard",
        "--directory",
      ],
      {cwd: repositoryRoot, encoding: "utf8", maxBuffer: 64 * 1024 * 1024},
    )
      .split("\n")
      .filter((line) => line !== "");
  } catch {
    return undefined;
  }
};

// validateHasherIgnore keeps hasher.ignorePatterns aligned with .gitignore.
// The glob walk does not apply .gitignore, so an ignored directory that no
// pattern covers starts being hashed, and a pattern that matches a declared
// input stops that task from re-running when the input changes.
const validateHasherIgnore = (
  repositoryRoot: string,
  projectFiles: readonly {path: string; text: string}[],
  check: Check,
): void => {
  const workspacePath = resolve(repositoryRoot, ".moon/workspace.yml");
  const failures = workspaceHasherFailures({
    workspaceText: existsSync(workspacePath)
      ? readFileSync(workspacePath, "utf8")
      : undefined,
    ignoredPaths: gitIgnoredPaths(repositoryRoot),
    projectFiles,
  });
  for (const failure of failures) check(false, failure);
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

const expectedPluginName = (pluginPackage: string): string => {
  const packageDirectory = basename(pluginPackage);
  return packageDirectory.startsWith("command-")
    ? `xonovex-${packageDirectory.slice("command-".length)}`
    : `xonovex-${packageDirectory}`;
};

const validateMarketplaceInventory = (
  label: string,
  marketplace: Marketplace,
  expectedNames: readonly string[],
  check: Check,
): void => {
  const counts = marketplace.plugins.reduce<ReadonlyMap<string, number>>(
    (current, entry) => {
      const next = new Map(current);
      next.set(entry.name, (next.get(entry.name) ?? 0) + 1);
      return next;
    },
    new Map(),
  );
  const expected = new Set(expectedNames);
  const missing = expectedNames.filter((name) => !counts.has(name)).toSorted();
  const unexpected = [...counts.keys()]
    .filter((name) => !expected.has(name))
    .toSorted();
  const duplicates = [...counts.entries()]
    .filter(([, count]) => count > 1)
    .map(([name, count]) => `${name} (${String(count)})`)
    .toSorted();

  check(
    missing.length === 0,
    `${label} is missing plugin entries: ${missing.join(", ")}`,
  );
  check(
    unexpected.length === 0,
    `${label} has unexpected plugin entries: ${unexpected.join(", ")}`,
  );
  check(
    duplicates.length === 0,
    `${label} has duplicate plugin entries: ${duplicates.join(", ")}`,
  );
};

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
  const packageLabel = relativePath(pluginPackage);
  const skillPackage = packageLabel.startsWith("packages/skill/");
  const packageJsonPath = resolve(pluginPackage, "package.json");
  const claudeManifestPath = resolve(
    pluginPackage,
    ".claude-plugin/plugin.json",
  );
  const codexManifestPath = resolve(pluginPackage, ".codex-plugin/plugin.json");

  check(existsSync(packageJsonPath), `${packageLabel} has package.json`);
  check(
    existsSync(claudeManifestPath),
    `${packageLabel} has a Claude plugin manifest`,
  );
  if (skillPackage) {
    check(
      existsSync(codexManifestPath),
      `${packageLabel} has a Codex plugin manifest`,
    );
  }
  if (
    !existsSync(packageJsonPath) ||
    !existsSync(claudeManifestPath) ||
    (skillPackage && !existsSync(codexManifestPath))
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
  const codexManifest = skillPackage
    ? readJson<PluginManifest>(
        codexManifestPath,
        relativePath(codexManifestPath),
        PluginManifestSchema,
        check,
      )
    : undefined;
  if (
    packageJson === undefined ||
    claudeManifest === undefined ||
    (skillPackage && codexManifest === undefined)
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
    claudeSources.get(claudeManifest.name) === source,
    `${claudeManifest.name} has the expected Claude marketplace source`,
  );

  if (codexManifest !== undefined) {
    check(
      codexManifest.version === releaseVersion,
      `${codexManifest.name} Codex manifest version matches ${releaseVersion}`,
    );
    check(
      claudeManifest.name === codexManifest.name,
      `${packageJson.name} manifest names match`,
    );
    check(
      codexSources.get(codexManifest.name) === source,
      `${codexManifest.name} has the expected Codex marketplace source`,
    );
    validateSkillPackaging(
      pluginPackage,
      packageLabel,
      claudeManifest,
      codexManifest,
      check,
    );
  }

  // package.json is the single description source: a plugin package must declare
  // one, and every manifest and marketplace entry must repeat it verbatim.
  check(
    packageJson.description !== undefined && packageJson.description !== "",
    `${packageJson.name} package declares a description`,
  );
  check(
    claudeManifest.description === packageJson.description,
    `${packageJson.name} package and Claude manifest descriptions match`,
  );
  check(
    claudeDescriptions.get(claudeManifest.name) === packageJson.description,
    `${packageJson.name} package and Claude marketplace descriptions match`,
  );
  if (codexManifest !== undefined) {
    check(
      codexManifest.description === packageJson.description,
      `${packageJson.name} package and Codex manifest descriptions match`,
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

  const commandPackages = childDirectories(
    resolve(repositoryRoot, "packages/command"),
    "packages/command",
    check,
  );
  const skillPackages = childDirectories(
    resolve(repositoryRoot, "packages/skill"),
    "packages/skill",
    check,
  );
  const pluginPackages = [...commandPackages, ...skillPackages];
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

  validateMarketplaceInventory(
    "Claude marketplace",
    marketplace,
    pluginPackages.map(expectedPluginName),
    check,
  );
  validateMarketplaceInventory(
    "Codex marketplace",
    codexMarketplace,
    skillPackages.map(expectedPluginName),
    check,
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

  const tagTaskDirectory = resolve(repositoryRoot, ".moon/tasks");
  const tagTaskFiles = existsSync(tagTaskDirectory)
    ? readdirSync(tagTaskDirectory)
        .toSorted()
        .filter((entry) => entry.endsWith(".yml"))
        .map((entry) => ({
          name: entry,
          text: readFileSync(resolve(tagTaskDirectory, entry), "utf8"),
        }))
    : [];
  check(tagTaskFiles.length > 0, ".moon/tasks must define tag task files");
  for (const failure of taskInheritanceFailures(tagTaskFiles)) {
    check(false, failure);
  }

  const packagesDirectory = resolve(repositoryRoot, "packages");
  const groupDirectories = childDirectories(
    packagesDirectory,
    "packages",
    check,
  );
  const projectDirectories = groupDirectories.flatMap((group) =>
    childDirectories(group, basename(group), check),
  );
  const projectFiles = projectDirectories
    .map((project) => resolve(project, "moon.yml"))
    .filter((path) => existsSync(path))
    .map((path) => ({
      path: path.slice(repositoryRoot.length + 1),
      text: readFileSync(path, "utf8"),
    }));
  check(projectFiles.length > 0, "packages must define project task files");
  for (const failure of coverageFloorFailures(projectFiles)) {
    check(false, failure);
  }

  validateHasherIgnore(repositoryRoot, projectFiles, check);

  // Every package group documents itself, and every AGENTS.md anywhere in the
  // repository carries the CLAUDE.md pointer that makes it load in Claude Code.
  const instructionDirectory = (directory: string): InstructionDirectory => {
    const claudePath = resolve(directory, "CLAUDE.md");
    return {
      path: directory.slice(repositoryRoot.length + 1) || ".",
      hasAgentsFile: existsSync(resolve(directory, "AGENTS.md")),
      claudeText: existsSync(claudePath)
        ? readFileSync(claudePath, "utf8")
        : undefined,
    };
  };
  const instructionDirectories = [
    repositoryRoot,
    resolve(repositoryRoot, ".moon"),
    ...groupDirectories,
    ...projectDirectories,
  ].map(instructionDirectory);
  const requiredInstructionDirectories = groupDirectories.map((group) =>
    group.slice(repositoryRoot.length + 1),
  );
  for (const failure of instructionDocFailures(
    instructionDirectories,
    requiredInstructionDirectories,
  )) {
    check(false, failure);
  }

  return {
    checks: checker.checks(),
    failures: checker.failures(),
    pluginPackages: pluginPackages.length,
  };
};
