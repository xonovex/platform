import {jsonInput} from "./file-input.js";
import type {MarketplaceEntries} from "./marketplace.js";
import {
  PackageManifestSchema,
  PluginManifestSchema,
  type PackageManifest,
  type PluginManifest,
} from "./release-inputs.js";
import type {PluginPackageSnapshot} from "./release-snapshot.js";
import type {Checker} from "./validate.js";

const SKILL_DIRECTORY = "./skills/";

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

const isSkillDirectoryOnly = (paths: readonly string[]): boolean =>
  paths.length === 1 && paths[0] === SKILL_DIRECTORY;

const validateSkillPackaging = (
  pluginPackage: PluginPackageSnapshot,
  claudeManifest: PluginManifest,
  codexManifest: PluginManifest,
  check: Checker["check"],
): void => {
  const label = pluginPackage.path;
  check(pluginPackage.skillCount > 0, `${label} contains skills`);
  const claudePaths = skillPaths(claudeManifest);
  const codexPaths = skillPaths(codexManifest);

  check(
    typeof claudeManifest.skills === "string" && claudePaths.length > 0,
    `${label} Claude manifest exposes a skill directory`,
  );
  check(
    typeof codexManifest.skills === "string",
    `${label} Codex manifest exposes a skill directory`,
  );
  check(
    isSkillDirectoryOnly(claudePaths),
    `${label} Claude manifest skill paths match ${SKILL_DIRECTORY}`,
  );
  check(
    isSkillDirectoryOnly(codexPaths),
    `${label} Codex manifest skill paths match ${SKILL_DIRECTORY}`,
  );
  for (const path of [...claudePaths, ...codexPaths]) {
    const direct = path === SKILL_DIRECTORY;
    check(direct, `${label} manifest skill path is direct: ${path}`);
    if (direct) {
      check(
        pluginPackage.skillsPathExists,
        `${label} manifest skill path resolves: ${path}`,
      );
    }
  }
};

interface PluginManifests {
  readonly packageJson: PackageManifest;
  readonly neutral: PluginManifest;
  readonly claude: PluginManifest;
  readonly codex: PluginManifest;
}

// Each manifest is checked for presence before any is parsed, so a package
// missing several reports every one of them.
const readManifests = (
  pluginPackage: PluginPackageSnapshot,
  check: Checker["check"],
): PluginManifests | undefined => {
  const label = pluginPackage.path;
  const {packageJson, neutralManifest, claudeManifest, codexManifest} =
    pluginPackage;
  check(packageJson !== undefined, `${label} has package.json`);
  check(
    neutralManifest !== undefined,
    `${label} has a neutral plugin manifest`,
  );
  check(claudeManifest !== undefined, `${label} has a Claude plugin manifest`);
  check(codexManifest !== undefined, `${label} has a Codex plugin manifest`);
  if (
    packageJson === undefined ||
    neutralManifest === undefined ||
    claudeManifest === undefined ||
    codexManifest === undefined
  ) {
    return undefined;
  }

  const manifests = {
    packageJson: jsonInput(
      packageJson,
      `${label}/package.json`,
      PackageManifestSchema,
      check,
    ),
    claude: jsonInput(
      claudeManifest,
      `${label}/.claude-plugin/plugin.json`,
      PluginManifestSchema,
      check,
    ),
    neutral: jsonInput(
      neutralManifest,
      `${label}/plugin.json`,
      PluginManifestSchema,
      check,
    ),
    codex: jsonInput(
      codexManifest,
      `${label}/.codex-plugin/plugin.json`,
      PluginManifestSchema,
      check,
    ),
  };
  return manifests.packageJson === undefined ||
    manifests.neutral === undefined ||
    manifests.claude === undefined ||
    manifests.codex === undefined
    ? undefined
    : {
        packageJson: manifests.packageJson,
        neutral: manifests.neutral,
        claude: manifests.claude,
        codex: manifests.codex,
      };
};

/**
 * Checks one plugin package against the release version and both marketplaces:
 * its manifests, their versions, names, and descriptions, and its skill layout.
 */
export const validatePluginPackage = (
  pluginPackage: PluginPackageSnapshot,
  releaseVersion: string,
  claudeMarketplace: MarketplaceEntries,
  codexMarketplace: MarketplaceEntries,
  check: Checker["check"],
): void => {
  const manifests = readManifests(pluginPackage, check);
  if (manifests === undefined) return;
  const {packageJson, neutral, claude, codex} = manifests;

  const source = `./${pluginPackage.path}`;
  check(
    neutral.name === claude.name &&
      neutral.version === releaseVersion &&
      neutral.description === packageJson.description &&
      neutral.skills === SKILL_DIRECTORY,
    `${packageJson.name} neutral manifest matches the plugin identity and skill directory`,
  );
  check(
    packageJson.version === releaseVersion,
    `${packageJson.name} package version matches ${releaseVersion}`,
  );
  check(
    claude.version === releaseVersion,
    `${claude.name} Claude manifest version matches ${releaseVersion}`,
  );
  check(
    claudeMarketplace.sources.get(claude.name) === source,
    `${claude.name} has the expected Claude marketplace source`,
  );
  check(
    codex.version === releaseVersion,
    `${codex.name} Codex manifest version matches ${releaseVersion}`,
  );
  check(claude.name === codex.name, `${packageJson.name} manifest names match`);
  check(
    codexMarketplace.sources.get(codex.name) === source,
    `${codex.name} has the expected Codex marketplace source`,
  );
  validateSkillPackaging(pluginPackage, claude, codex, check);

  // package.json is the single description source: a plugin package must declare
  // one, and every manifest and marketplace entry must repeat it verbatim.
  check(
    packageJson.description !== undefined && packageJson.description !== "",
    `${packageJson.name} package declares a description`,
  );
  check(
    claude.description === packageJson.description,
    `${packageJson.name} package and Claude manifest descriptions match`,
  );
  check(
    claudeMarketplace.descriptions.get(claude.name) === packageJson.description,
    `${packageJson.name} package and Claude marketplace descriptions match`,
  );
  check(
    codex.description === packageJson.description,
    `${packageJson.name} package and Codex manifest descriptions match`,
  );
  check(
    codexMarketplace.descriptions.get(codex.name) === packageJson.description,
    `${packageJson.name} package and Codex marketplace descriptions match`,
  );

  for (const [dependency, version] of dependencyEntries(packageJson)) {
    if (dependency.startsWith("@xonovex/plugin-")) {
      check(
        version === releaseVersion,
        `${packageJson.name} dependency ${dependency} matches ${releaseVersion}`,
      );
    }
  }
};
