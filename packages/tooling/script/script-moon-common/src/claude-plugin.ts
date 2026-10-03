import {cpSync, mkdirSync, writeFileSync} from "node:fs";
import {basename, join} from "node:path";
import {nodeFileSystem, type FileSystem} from "./file-system.js";
import {pluginSkillDirectories} from "./package-layout.js";

export interface ClaudePluginStage {
  readonly directory: string;
  readonly name: string;
  readonly guides: readonly string[];
}

export const planClaudePluginStages = (
  args: readonly string[],
  workspace: string,
  selectedSkills: readonly string[],
  fs: FileSystem = nodeFileSystem,
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
    const guides = pluginSkillDirectories(argument, fs);
    const selected = guides.filter((guide) =>
      selectedSkills.includes(basename(guide)),
    );
    const directory = join(workspace, `plugin-${String(index)}`);
    stages.push({
      directory,
      name: parsed.name,
      guides: selected.length > 0 ? selected : guides,
    });
    return directory;
  });
  return {args: stagedArgs, stages};
};

export const stageClaudePluginArguments = (
  args: readonly string[],
  workspace: string,
  selectedSkills: readonly string[],
): readonly string[] => {
  const plan = planClaudePluginStages(args, workspace, selectedSkills);
  for (const stage of plan.stages) {
    mkdirSync(join(stage.directory, ".claude-plugin"), {recursive: true});
    for (const guide of stage.guides)
      cpSync(guide, join(stage.directory, "skills", basename(guide)), {
        recursive: true,
      });
    writeFileSync(
      join(stage.directory, ".claude-plugin", "plugin.json"),
      JSON.stringify({
        name: stage.name,
        skills: stage.guides.map((guide) => `./skills/${basename(guide)}`),
      }),
    );
  }
  return plan.args;
};
