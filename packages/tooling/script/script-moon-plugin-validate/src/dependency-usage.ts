import {basename, dirname, join, relative} from "node:path";
import {
  nodeFileSystem,
  type FileSystem,
} from "@xonovex/script-moon-common/file-system";
import {pluginSkillDirectories} from "@xonovex/script-moon-common/package-layout";
import {issue, type ValidationIssue} from "./validation.js";

const SKILL_PLUGIN_PREFIX = "xonovex-";
const SKILL_PACKAGE_PREFIX = "@xonovex/plugin-";

const isBoundary = (character: string | undefined): boolean =>
  character === undefined || !/[\p{Letter}\p{Number}-]/u.test(character);

// The name is matched as a whole word so a longer skill name that contains it,
// such as `using-github` for `using-git`, is not read as a mention.
const namesSkill = (source: string, skill: string): boolean => {
  for (
    let index = source.indexOf(skill);
    index !== -1;
    index = source.indexOf(skill, index + 1)
  ) {
    if (
      isBoundary(source[index - 1]) &&
      isBoundary(source[index + skill.length])
    ) {
      return true;
    }
  }
  return false;
};

export interface DependencyUsageInput {
  readonly packageDirectory: string;
  readonly repositoryRoot: string;
  readonly pluginDependencies: readonly string[];
  readonly packageDependencies: readonly string[];
}

// checkDependencyUsage requires a reason for every declared skill dependency:
// a plugin installs another plugin only because one of its skills hands work
// to a skill of that plugin, so a SKILL.md of this plugin must name one.
export const checkDependencyUsage = (
  input: DependencyUsageInput,
  fs: FileSystem = nodeFileSystem,
): readonly ValidationIssue[] => {
  const sources = pluginSkillDirectories(input.packageDirectory, fs).map(
    (directory) => fs.readText(join(directory, "SKILL.md")),
  );
  const isUnused = (plugin: string): boolean => {
    const dependencyDirectory = join(
      dirname(input.packageDirectory),
      `plugin-${plugin.slice(SKILL_PLUGIN_PREFIX.length)}`,
    );
    const skills = pluginSkillDirectories(dependencyDirectory, fs).map(
      (directory) => basename(directory),
    );
    return skills.every((skill) =>
      sources.every((source) => !namesSkill(source, skill)),
    );
  };
  const unusedReason =
    "is declared but no skill of this plugin names one of its skills";

  return [
    ...input.pluginDependencies
      .filter(
        (plugin) => plugin.startsWith(SKILL_PLUGIN_PREFIX) && isUnused(plugin),
      )
      .map((plugin) =>
        issue(
          "dependency.plugin-unused",
          relative(
            input.repositoryRoot,
            join(input.packageDirectory, ".claude-plugin", "plugin.json"),
          ),
          `plugin dependency '${plugin}' ${unusedReason}`,
        ),
      ),
    ...input.packageDependencies
      .filter(
        (name) =>
          name.startsWith(SKILL_PACKAGE_PREFIX) &&
          isUnused(
            `${SKILL_PLUGIN_PREFIX}${name.slice(SKILL_PACKAGE_PREFIX.length)}`,
          ),
      )
      .map((name) =>
        issue(
          "dependency.package-unused",
          relative(
            input.repositoryRoot,
            join(input.packageDirectory, "package.json"),
          ),
          `package dependency '${name}' ${unusedReason}`,
        ),
      ),
  ];
};
