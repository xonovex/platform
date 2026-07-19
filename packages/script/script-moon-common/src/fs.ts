import {readdirSync, readFileSync, statSync} from "node:fs";
import {dirname, join, resolve} from "node:path";

export const isFile = (path: string): boolean => {
  try {
    return statSync(path).isFile();
  } catch {
    return false;
  }
};

export const isDirectory = (path: string): boolean => {
  try {
    return statSync(path).isDirectory();
  } catch {
    return false;
  }
};

export const resolveGuideDirectory = (base: string): string => {
  if (isFile(join(base, "SKILL.md"))) return base;
  const nested = readdirSync(base)
    .map((entry) => join(base, entry))
    .filter((path) => isFile(join(path, "SKILL.md")));
  if (nested.length > 1) {
    throw new Error(
      `multiple SKILL.md found under ${base}; pass one explicitly`,
    );
  }
  return nested[0] ?? base;
};

interface ClaudePluginManifest {
  readonly name: string;
  readonly dependencies: readonly string[];
}

const CLAUDE_SKILL_PLUGIN_FIELDS = new Set([
  "author",
  "dependencies",
  "description",
  "name",
  "skills",
  "version",
]);

const readClaudePluginManifest = (directory: string): ClaudePluginManifest => {
  const manifestPath = join(directory, ".claude-plugin", "plugin.json");
  let parsed: unknown;
  try {
    parsed = JSON.parse(readFileSync(manifestPath, "utf8"));
  } catch {
    throw new Error(`invalid Claude plugin manifest: ${manifestPath}`);
  }
  if (typeof parsed !== "object" || parsed === null || Array.isArray(parsed)) {
    throw new Error(`invalid Claude plugin manifest: ${manifestPath}`);
  }
  const record = parsed as Record<string, unknown>;
  const unsupportedField = Object.keys(record).find(
    (field) => !CLAUDE_SKILL_PLUGIN_FIELDS.has(field),
  );
  if (unsupportedField !== undefined) {
    throw new Error(
      `Claude eval plugin is not skill-only: ${manifestPath} has ${unsupportedField}`,
    );
  }
  if (typeof record.name !== "string" || record.name.length === 0) {
    throw new Error(`Claude plugin manifest has no name: ${manifestPath}`);
  }
  const dependencies = record.dependencies ?? [];
  if (
    !Array.isArray(dependencies) ||
    dependencies.some((dependency) => typeof dependency !== "string")
  ) {
    throw new Error(`Claude plugin dependencies are invalid: ${manifestPath}`);
  }
  if (
    !Array.isArray(record.skills) ||
    record.skills.length === 0 ||
    record.skills.some((skill) => typeof skill !== "string")
  ) {
    throw new Error(`Claude plugin skills are invalid: ${manifestPath}`);
  }
  const executableComponents = [
    "commands",
    "agents",
    "hooks",
    ".mcp.json",
    ".lsp.json",
    "settings.json",
    join(".claude", "settings.json"),
  ];
  const executableComponent = executableComponents.find((entry) => {
    const path = join(directory, entry);
    return isFile(path) || isDirectory(path);
  });
  if (executableComponent !== undefined) {
    throw new Error(
      `Claude eval plugin is not skill-only: ${directory} contains ${executableComponent}`,
    );
  }
  return {name: record.name, dependencies};
};

export const resolveClaudePluginDirectories = (
  targetDirectory: string,
): readonly string[] => {
  const target = resolve(targetDirectory);
  const siblingRoot = dirname(target);
  const directoryByName = new Map<string, string>();
  for (const entry of readdirSync(siblingRoot)) {
    const directory = join(siblingRoot, entry);
    if (!isFile(join(directory, ".claude-plugin", "plugin.json"))) continue;
    const manifest = readClaudePluginManifest(directory);
    directoryByName.set(manifest.name, directory);
  }

  const ordered: string[] = [];
  const visited = new Set<string>();
  const visit = (directory: string): void => {
    const manifest = readClaudePluginManifest(directory);
    if (visited.has(manifest.name)) return;
    visited.add(manifest.name);
    for (const dependency of manifest.dependencies) {
      const dependencyDirectory = directoryByName.get(dependency);
      if (dependencyDirectory === undefined) {
        throw new Error(
          `local Claude plugin dependency not found: ${manifest.name} -> ${dependency}`,
        );
      }
      visit(dependencyDirectory);
    }
    ordered.push(directory);
  };
  visit(target);
  return ordered;
};
