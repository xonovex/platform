import {existsSync, readdirSync, readFileSync, statSync} from "node:fs";
import {basename, dirname, join, resolve} from "node:path";
import {
  parseInstalledSkillInventory,
  type InstalledSkill,
  type ParseResult,
} from "./skill-composition-contract.js";

const IGNORED_DIRECTORIES = new Set([
  ".git",
  ".moon",
  "coverage",
  "dist",
  "node_modules",
  "target",
]);

interface DiscoveredPlugin {
  readonly dependencies: readonly string[];
  readonly guides: readonly string[];
  readonly packageDependencies: Readonly<Record<string, string>>;
  readonly packageName: string | undefined;
  readonly packagePath: string;
  readonly plugin: string;
  readonly version: string;
}

const isRecord = (value: unknown): value is Readonly<Record<string, unknown>> =>
  typeof value === "object" && value !== null && !Array.isArray(value);

const readJson = (path: string): unknown => {
  try {
    return JSON.parse(readFileSync(path, "utf8")) as unknown;
  } catch (error) {
    throw new Error(`cannot parse JSON at ${path}`, {cause: error});
  }
};

const stringArray = (value: unknown): readonly string[] =>
  Array.isArray(value) && value.every((entry) => typeof entry === "string")
    ? value
    : [];

const skillPaths = (
  value: unknown,
  manifestPath: string,
): readonly string[] => {
  if (value === undefined) return [];
  if (typeof value === "string") return [value];
  const values = stringArray(value);
  if (values.length === (Array.isArray(value) ? value.length : -1)) {
    return values;
  }
  throw new TypeError(
    `${manifestPath}: skills must be a string or string array`,
  );
};

const manifestPathsBelow = (root: string): readonly string[] => {
  if (!existsSync(root)) return [];
  if (!statSync(root).isDirectory()) return [];
  const directClaude = join(root, ".claude-plugin", "plugin.json");
  const directCodex = join(root, ".codex-plugin", "plugin.json");
  if (existsSync(directClaude)) return [directClaude];
  if (existsSync(directCodex)) return [directCodex];
  return readdirSync(root, {withFileTypes: true}).flatMap((entry) => {
    if (!entry.isDirectory() || IGNORED_DIRECTORIES.has(entry.name)) return [];
    return manifestPathsBelow(join(root, entry.name));
  });
};

const readPlugin = (manifestPath: string): DiscoveredPlugin | undefined => {
  const pluginRoot = dirname(dirname(manifestPath));
  const pluginJson = readJson(manifestPath);
  const packagePath = join(pluginRoot, "package.json");
  const packageJson = existsSync(packagePath) ? readJson(packagePath) : {};
  if (!isRecord(pluginJson) || !isRecord(packageJson)) return undefined;
  const plugin = pluginJson.name;
  const version = pluginJson.version;
  if (typeof plugin !== "string" || typeof version !== "string") {
    throw new TypeError(
      `${manifestPath}: plugin name and version are required`,
    );
  }
  if (
    pluginJson.dependencies !== undefined &&
    stringArray(pluginJson.dependencies).length !==
      (Array.isArray(pluginJson.dependencies)
        ? pluginJson.dependencies.length
        : -1)
  ) {
    throw new TypeError(`${manifestPath}: dependencies must be a string array`);
  }
  const rawPackageDependencies = packageJson.dependencies;
  if (
    rawPackageDependencies !== undefined &&
    (!isRecord(rawPackageDependencies) ||
      Object.values(rawPackageDependencies).some(
        (value) => typeof value !== "string",
      ))
  ) {
    throw new TypeError(
      `${packagePath}: dependencies must map package names to versions`,
    );
  }
  const packageDependencies = (rawPackageDependencies ?? {}) as Readonly<
    Record<string, string>
  >;
  const packageName =
    typeof packageJson.name === "string" ? packageJson.name : undefined;
  return {
    dependencies: stringArray(pluginJson.dependencies),
    guides: skillPaths(pluginJson.skills, manifestPath).map((path) =>
      basename(path.replace(/\/$/u, "")),
    ),
    packageDependencies,
    packageName,
    packagePath: pluginRoot,
    plugin,
    version,
  };
};

const dependencyVersion = (
  owner: DiscoveredPlugin,
  target: DiscoveredPlugin | undefined,
): string | undefined => {
  if (target?.packageName === undefined) return undefined;
  return owner.packageDependencies[target.packageName];
};

export const discoverInstalledSkillInventory = (
  roots: readonly string[],
): ParseResult<readonly InstalledSkill[]> => {
  const manifestPaths = [
    ...new Set(roots.flatMap((root) => manifestPathsBelow(resolve(root)))),
  ].toSorted();
  const errors: string[] = [];
  const plugins = manifestPaths.flatMap((path) => {
    try {
      const plugin = readPlugin(path);
      return plugin === undefined ? [] : [plugin];
    } catch (error) {
      errors.push(error instanceof Error ? error.message : String(error));
      return [];
    }
  });
  if (errors.length > 0) return {success: false, errors};
  const pluginByName = new Map(
    plugins.map((plugin) => [plugin.plugin, plugin]),
  );
  const input = plugins.flatMap((plugin) =>
    plugin.guides.map((guide) => ({
      guide,
      implementationVersion: plugin.version,
      dependencies: plugin.dependencies.map((dependencyPlugin) => {
        const version = dependencyVersion(
          plugin,
          pluginByName.get(dependencyPlugin),
        );
        return {
          plugin: dependencyPlugin,
          ...(version === undefined ? {} : {implementationVersion: version}),
        };
      }),
      packagePath: plugin.packagePath,
      plugin: plugin.plugin,
      sourcesPath: join(plugin.packagePath, guide, "SOURCES.md"),
    })),
  );
  return parseInstalledSkillInventory(input);
};
