import {join, relative} from "node:path";
import {isRecord} from "@xonovex/script-moon-common/records";
import {
  fail,
  pass,
  type Finding,
} from "@xonovex/script-moon-skill-catalog-common/finding";
import {
  pluginRoot,
  type MarkdownDocument,
  type PluginPackage,
} from "./catalog-scan.js";
import {dependencyCycles} from "./dependency-graph.js";
import {missingFile, parseJsonFile, type FileRead} from "./file-read.js";
import type {SkillPluginManifest} from "./plugin-manifest.js";
import {checkReferenceOverlap} from "./reference-overlap.js";
import {checkSkillPackage} from "./skill-packaging.js";

const NPM_PLUGIN_PREFIX = "@xonovex/plugin-";

interface CheckedPlugin {
  readonly manifest: SkillPluginManifest;
  /** Reference files with paths relative to the plugin package. */
  readonly references: readonly MarkdownDocument[];
}

// A plugin named xonovex-<x> lives in plugin-<x> and is published as
// @xonovex/plugin-<x>.
const packageDirectoryName = (pluginName: string): string =>
  pluginName.replace(/^xonovex-/u, "plugin-");

const isStringRecord = (
  value: unknown,
): value is Readonly<Record<string, string>> =>
  isRecord(value) &&
  Object.values(value).every((entry) => typeof entry === "string");

interface PluginContext {
  readonly repoRoot: string;
  readonly packageFiles: ReadonlyMap<string, FileRead>;
  readonly plugins: ReadonlyMap<string, CheckedPlugin>;
}

const packageFile = (
  pluginName: string,
  context: PluginContext,
): {readonly displayPath: string; readonly file: FileRead} => {
  const directoryName = packageDirectoryName(pluginName);
  const path = join(
    pluginRoot(context.repoRoot),
    directoryName,
    "package.json",
  );
  return {
    displayPath: relative(context.repoRoot, path),
    file: context.packageFiles.get(directoryName) ?? missingFile(path),
  };
};

type PackageDependencies =
  | {
      readonly kind: "read";
      readonly dependencies: Readonly<Record<string, string>>;
    }
  | {readonly kind: "invalid"; readonly finding: Finding};

const readPackageDependencies = (
  pluginName: string,
  context: PluginContext,
): PackageDependencies => {
  const {displayPath, file} = packageFile(pluginName, context);
  const parsed = parseJsonFile(file);
  if (parsed.kind === "invalid") {
    return {
      kind: "invalid",
      finding: fail(
        `skill dependencies: cannot read ${displayPath}: ${parsed.message}`,
      ),
    };
  }
  const {value} = parsed;
  if (isRecord(value) && isStringRecord(value.dependencies)) {
    return {kind: "read", dependencies: value.dependencies};
  }
  if (isRecord(value) && value.dependencies === undefined) {
    return {kind: "read", dependencies: {}};
  }
  return {
    kind: "invalid",
    finding: fail(
      `skill dependencies: ${displayPath} dependencies must map package names to versions`,
    ),
  };
};

// A dependent pins the exact version the dependency's own package.json declares.
const checkDependencyVersion = (
  dependentName: string,
  dependency: string,
  declaredVersion: string,
  context: PluginContext,
): readonly Finding[] => {
  const {displayPath, file} = packageFile(dependency, context);
  const parsed = parseJsonFile(file);
  if (parsed.kind === "invalid") {
    return [
      fail(`skill dependencies: cannot read ${displayPath}: ${parsed.message}`),
    ];
  }
  const {value} = parsed;
  if (!isRecord(value) || typeof value.version !== "string") {
    return [fail(`skill dependencies: ${displayPath} needs a string version`)];
  }
  return declaredVersion === value.version
    ? []
    : [
        fail(
          `skill dependencies: ${dependentName} pins @xonovex/${packageDirectoryName(dependency)}@${declaredVersion}; expected exact installed version ${value.version}`,
        ),
      ];
};

// One plugin-manifest dependency resolves, is pinned in package.json, and
// shares no reference surface with its dependent.
const checkManifestDependency = (
  plugin: CheckedPlugin,
  dependency: string,
  packageDependencies: Readonly<Record<string, string>> | undefined,
  context: PluginContext,
): readonly Finding[] => {
  const {name} = plugin.manifest;
  const target = context.plugins.get(dependency);
  if (target === undefined) {
    return [
      fail(`skill dependencies: ${name} depends on missing ${dependency}`),
    ];
  }
  const npmDependency = `@xonovex/${packageDirectoryName(dependency)}`;
  const declaredVersion = packageDependencies?.[npmDependency];
  return [
    ...(declaredVersion === undefined
      ? [
          fail(
            `skill dependencies: ${name} declares ${dependency} in its plugin manifests but omits ${npmDependency} from package.json`,
          ),
        ]
      : checkDependencyVersion(name, dependency, declaredVersion, context)),
    ...checkReferenceOverlap(
      {name, references: plugin.references},
      {name: dependency, references: target.references},
    ),
  ];
};

const checkDeclaredDependencies = (
  plugin: CheckedPlugin,
  context: PluginContext,
): readonly Finding[] => {
  const {manifest} = plugin;
  const packageDependencies = readPackageDependencies(manifest.name, context);
  const dependencies =
    packageDependencies.kind === "read"
      ? packageDependencies.dependencies
      : undefined;
  return [
    ...(packageDependencies.kind === "invalid"
      ? [packageDependencies.finding]
      : []),
    ...Object.keys(dependencies ?? {})
      .filter((dependency) => dependency.startsWith(NPM_PLUGIN_PREFIX))
      .flatMap((npmDependency) => {
        const pluginDependency = `xonovex-${npmDependency.slice(NPM_PLUGIN_PREFIX.length)}`;
        return manifest.dependencies.includes(pluginDependency)
          ? []
          : [
              fail(
                `skill dependencies: ${manifest.name} declares ${npmDependency} in package.json but omits ${pluginDependency} from its plugin manifests`,
              ),
            ];
      }),
    ...manifest.dependencies.flatMap((dependency) =>
      checkManifestDependency(plugin, dependency, dependencies, context),
    ),
  ];
};

/**
 * Checks the plugin packages: each ships an agreeing manifest pair, declares
 * the same hard dependencies to npm and to the harness, pins them exactly,
 * copies no dependency reference, and forms no cycle.
 */
export const checkSkillDependencies = (
  packages: readonly PluginPackage[],
  repoRoot: string,
): readonly Finding[] => {
  const findings: Finding[] = [];
  const plugins = new Map<string, CheckedPlugin>();
  let pairs = 0;
  for (const plugin of packages.filter(({name}) =>
    name.startsWith("plugin-"),
  )) {
    const checked = checkSkillPackage(plugin, repoRoot);
    if (checked === undefined) continue;
    findings.push(...checked.findings);
    if (checked.pair) pairs += 1;
    if (checked.manifest === undefined) continue;
    if (plugins.has(checked.manifest.name)) {
      findings.push(
        fail(
          `skill dependencies: duplicate plugin name ${checked.manifest.name}`,
        ),
      );
      continue;
    }
    plugins.set(checked.manifest.name, {
      manifest: checked.manifest,
      references: plugin.skills.flatMap(({references}) =>
        references.map(({path, text}) => ({
          path: relative(plugin.directory, path),
          text,
        })),
      ),
    });
  }

  const context: PluginContext = {
    repoRoot,
    packageFiles: new Map(
      packages.map(({name, packageJson}) => [name, packageJson]),
    ),
    plugins,
  };
  findings.push(
    ...[...plugins.values()].flatMap((plugin) =>
      checkDeclaredDependencies(plugin, context),
    ),
    ...[
      ...dependencyCycles(
        new Map(
          [...plugins].map(([name, {manifest}]) => [
            name,
            manifest.dependencies,
          ]),
        ),
      ),
    ].map((cycle) => fail(`skill dependencies: dependency cycle ${cycle}`)),
  );

  return findings.length === 0 && pairs > 0
    ? [
        pass(
          `skill dependencies: ${String(pairs)} manifest pair(s) agree with no dangling dependencies or cycles`,
        ),
      ]
    : findings;
};
