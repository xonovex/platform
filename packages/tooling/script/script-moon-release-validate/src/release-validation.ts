import {basename, dirname} from "node:path";
import {
  binEntries,
  binTargetFailures,
  type PackageBins,
} from "./bin-targets.js";
import {coverageFloorFailures} from "./coverage-floors.js";
import {dependencySourceFailures} from "./dependency-sources.js";
import {jsonInput, textInput} from "./file-input.js";
import {workspaceHasherFailures} from "./hasher-ignore.js";
import {instructionDocFailures} from "./instruction-docs.js";
import {
  expectedPluginName,
  marketplaceEntries,
  validateMarketplaceInventory,
} from "./marketplace.js";
import {validatePluginPackage} from "./plugin-package.js";
import {
  LockfileSchema,
  MarketplaceSchema,
  PackageManifestSchema,
  type Lockfile,
  type Marketplace,
} from "./release-inputs.js";
import type {
  PluginPackageSnapshot,
  ReleaseSnapshot,
} from "./release-snapshot.js";
import {taskInheritanceFailures} from "./task-inheritance.js";
import {
  createChecker,
  forbiddenClaims,
  markdownLinkTargets,
  releaseWorkflowFailures,
  type Checker,
} from "./validate.js";

export interface ReleaseValidationResult {
  readonly checks: number;
  readonly failures: readonly string[];
  readonly pluginPackages: number;
}

type Check = Checker["check"];

const reportAll = (failures: readonly string[], check: Check): void => {
  for (const failure of failures) check(false, failure);
};

const validateLockfile = (
  releaseVersion: string,
  pluginPackages: readonly PluginPackageSnapshot[],
  lockfile: Lockfile,
  check: Check,
): void => {
  for (const pluginPackage of pluginPackages) {
    check(
      lockfile.packages?.[pluginPackage.path]?.version === releaseVersion,
      `package-lock records ${pluginPackage.path}@${releaseVersion}`,
    );
  }
};

// The plugin packages move in lockstep: one release version across both
// marketplaces, every plugin package, and the lockfile.
const validatePluginRelease = (
  snapshot: ReleaseSnapshot,
  claudeMarketplace: Marketplace,
  codexMarketplace: Marketplace,
  check: Check,
): readonly PluginPackageSnapshot[] => {
  const releaseVersion = claudeMarketplace.metadata?.version ?? "";
  check(
    /^\d+\.\d+\.\d+$/u.test(releaseVersion),
    "marketplace release must use a semantic version",
  );

  const pluginPackages =
    snapshot.pluginPackages.kind === "listed"
      ? snapshot.pluginPackages.packages
      : [];
  if (snapshot.pluginPackages.kind === "unlisted") {
    check(
      false,
      `packages/runtime/plugin could not be enumerated: ${snapshot.pluginPackages.reason}`,
    );
  }
  const claudeEntries = marketplaceEntries(claudeMarketplace);
  const codexEntries = marketplaceEntries(codexMarketplace);
  for (const pluginPackage of pluginPackages) {
    validatePluginPackage(
      pluginPackage,
      releaseVersion,
      claudeEntries,
      codexEntries,
      check,
    );
  }

  const expectedNames = pluginPackages.map((pluginPackage) =>
    expectedPluginName(pluginPackage.path),
  );
  validateMarketplaceInventory(
    "Claude marketplace",
    claudeMarketplace,
    expectedNames,
    check,
  );
  validateMarketplaceInventory(
    "Codex marketplace",
    codexMarketplace,
    expectedNames,
    check,
  );

  const lockfile = jsonInput(
    snapshot.lockfile,
    "package-lock.json",
    LockfileSchema,
    check,
  );
  if (lockfile !== undefined) {
    validateLockfile(releaseVersion, pluginPackages, lockfile, check);
  }
  return pluginPackages;
};

const validateReadme = (snapshot: ReleaseSnapshot, check: Check): void => {
  const readme = textInput(snapshot.readme, "README.md", check);
  if (readme === undefined) return;
  for (const target of markdownLinkTargets(readme)) {
    check(
      snapshot.existingLinkTargets.has(target),
      `README.md link resolves: ${target}`,
    );
  }
  for (const forbiddenClaim of forbiddenClaims) {
    check(
      !forbiddenClaim.test(readme),
      `root README rejects ${String(forbiddenClaim)}`,
    );
  }
};

const validateReleaseWorkflow = (
  snapshot: ReleaseSnapshot,
  check: Check,
): void => {
  const workflow = textInput(
    snapshot.releaseWorkflow,
    ".github/workflows/release.yml",
    check,
  );
  if (workflow !== undefined) {
    reportAll(releaseWorkflowFailures(workflow), check);
  }
};

const validateTaskFiles = (snapshot: ReleaseSnapshot, check: Check): void => {
  check(
    snapshot.tagTaskFiles.length > 0,
    ".moon/tasks must define tag task files",
  );
  reportAll(taskInheritanceFailures(snapshot.tagTaskFiles), check);

  const projectFiles = snapshot.projects.map(({path, text}) => ({path, text}));
  check(projectFiles.length > 0, "packages must define project task files");
  reportAll(coverageFloorFailures(projectFiles), check);

  // A project's id is its directory name and its source is the directory itself,
  // which is how a dependency named in dependsOn resolves to the src glob its
  // dependents must declare.
  reportAll(
    dependencySourceFailures(
      snapshot.projects.map((project) => ({
        id: basename(dirname(project.path)),
        source: dirname(project.path),
        text: project.text,
        packageName:
          project.packageJsonText === undefined
            ? undefined
            : (JSON.parse(project.packageJsonText) as {name?: string}).name,
        configTexts: project.configTexts,
      })),
    ),
    check,
  );

  // hasher.ignorePatterns must follow .gitignore. The glob walk does not apply
  // .gitignore, so an ignored directory that no pattern covers starts being
  // hashed, and a pattern that matches a declared input stops that task from
  // re-running when the input changes.
  reportAll(
    workspaceHasherFailures({
      workspaceText: snapshot.workspaceText,
      ignoredPaths: snapshot.ignoredPaths,
      projectFiles,
    }),
    check,
  );
};

// Every workspace bin must stay linkable on a cold install.
const validateBinTargets = (snapshot: ReleaseSnapshot, check: Check): void => {
  if (snapshot.trackedPaths === undefined) {
    check(false, "git could not list the committed files");
    return;
  }
  const packages: PackageBins[] = [];
  for (const {directory, manifest: read} of snapshot.packageManifests) {
    const manifest = jsonInput(
      read,
      `${directory}/package.json`,
      PackageManifestSchema,
      check,
    );
    if (manifest?.bin === undefined) continue;
    packages.push({
      manifestPath: `${directory}/package.json`,
      packageDirectory: directory,
      isPrivate: manifest.private === true,
      files: manifest.files,
      bins: binEntries(manifest.bin, manifest.name),
    });
  }
  reportAll(
    binTargetFailures({packages, trackedPaths: snapshot.trackedPaths}),
    check,
  );
};

/** Checks every release input the snapshot holds. */
export const validateRelease = (
  snapshot: ReleaseSnapshot,
): ReleaseValidationResult => {
  const checker = createChecker();
  const check = checker.check;
  const result = (pluginPackages: number): ReleaseValidationResult => ({
    checks: checker.checks(),
    failures: checker.failures(),
    pluginPackages,
  });

  const claudeMarketplace = jsonInput(
    snapshot.claudeMarketplace,
    ".claude-plugin/marketplace.json",
    MarketplaceSchema,
    check,
  );
  const codexMarketplace = jsonInput(
    snapshot.codexMarketplace,
    ".agents/plugins/marketplace.json",
    MarketplaceSchema,
    check,
  );
  if (claudeMarketplace === undefined || codexMarketplace === undefined) {
    return result(0);
  }

  const pluginPackages = validatePluginRelease(
    snapshot,
    claudeMarketplace,
    codexMarketplace,
    check,
  );
  validateReadme(snapshot, check);
  validateReleaseWorkflow(snapshot, check);
  validateTaskFiles(snapshot, check);
  validateBinTargets(snapshot, check);

  // Every package group documents itself, and every AGENTS.md anywhere in the
  // repository carries the CLAUDE.md pointer that makes it load in Claude Code.
  reportAll(
    instructionDocFailures(
      snapshot.instructionDirectories,
      snapshot.requiredInstructionDirectories,
    ),
    check,
  );

  return result(pluginPackages.length);
};
