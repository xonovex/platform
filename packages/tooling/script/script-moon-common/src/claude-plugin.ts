import {basename, dirname, join, resolve} from "node:path";
import type {FileSystem} from "./file-system.js";
import {pluginSkillDirectories} from "./package-layout.js";

export interface ClaudePluginStage {
  readonly directory: string;
  readonly name: string;
  readonly skillDirectories: readonly string[];
}

export const planClaudePluginStages = (
  args: readonly string[],
  workspace: string,
  selectedSkills: readonly string[],
  fs: FileSystem,
): {
  readonly args: readonly string[];
  readonly stages: readonly ClaudePluginStage[];
} => {
  const stages: ClaudePluginStage[] = [];
  const stagedArgs = args.map((argument, index) => {
    if (args[index - 1] !== "--plugin-dir") return argument;
    const parsed = JSON.parse(
      fs.readText(join(argument, ".claude-plugin", "plugin.json")),
    ) as {readonly name: string; readonly skills: unknown};
    if (parsed.skills !== "./skills/") return argument;
    const available = pluginSkillDirectories(argument, fs);
    const selected = available.filter((skillDirectory) =>
      selectedSkills.includes(basename(skillDirectory)),
    );
    const directory = join(workspace, `plugin-${String(index)}`);
    stages.push({
      directory,
      name: parsed.name,
      skillDirectories: selected.length > 0 ? selected : available,
    });
    return directory;
  });
  return {args: stagedArgs, stages};
};

export const stageClaudePluginArguments = (
  args: readonly string[],
  workspace: string,
  selectedSkills: readonly string[],
  fs: FileSystem,
): readonly string[] => {
  const plan = planClaudePluginStages(args, workspace, selectedSkills, fs);
  for (const stage of plan.stages) {
    for (const skillDirectory of stage.skillDirectories)
      fs.copyDirectory(
        skillDirectory,
        join(stage.directory, "skills", basename(skillDirectory)),
      );
    fs.writeFile(
      join(stage.directory, ".claude-plugin", "plugin.json"),
      JSON.stringify({
        name: stage.name,
        skills: stage.skillDirectories.map(
          (skillDirectory) => `./skills/${basename(skillDirectory)}`,
        ),
      }),
    );
  }
  return plan.args;
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

const readClaudePluginManifest = (
  directory: string,
  fs: FileSystem,
): ClaudePluginManifest => {
  const manifestPath = join(directory, ".claude-plugin", "plugin.json");
  let parsed: unknown;
  try {
    parsed = JSON.parse(fs.readText(manifestPath));
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
  const grouped = record.skills === "./skills/";
  if (
    !grouped &&
    (!Array.isArray(record.skills) ||
      record.skills.length === 0 ||
      record.skills.some((skill) => typeof skill !== "string"))
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
    return fs.isFile(path) || fs.isDirectory(path);
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
  fs: FileSystem,
): readonly string[] => {
  const target = resolve(targetDirectory);
  const siblingRoot = dirname(target);
  const directoryByName = new Map<string, string>();
  for (const entry of fs.readDirectory(siblingRoot)) {
    const directory = join(siblingRoot, entry);
    if (!fs.isFile(join(directory, ".claude-plugin", "plugin.json"))) continue;
    const manifest = readClaudePluginManifest(directory, fs);
    directoryByName.set(manifest.name, directory);
  }

  const ordered: string[] = [];
  const visited = new Set<string>();
  const visit = (directory: string): void => {
    const manifest = readClaudePluginManifest(directory, fs);
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
