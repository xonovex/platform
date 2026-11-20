import {basename, dirname, join, relative} from "node:path";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {pluginSkillDirectories} from "@xonovex/script-moon-common/package-layout";
import {pluginHandoffNames} from "@xonovex/script-moon-common/skill-handoffs";
import {issue, type ValidationIssue} from "./validation.js";

const SKILL_PLUGIN_PREFIX = "xonovex-";
const SKILL_PACKAGE_PREFIX = "@xonovex/plugin-";

export interface DependencyUsageInput {
  readonly packageDirectory: string;
  readonly repositoryRoot: string;
  readonly pluginDependencies: readonly string[];
  readonly packageDependencies: readonly string[];
}

// checkDependencyUsage requires a reason for every declared skill dependency:
// a plugin installs another plugin only because one of its skills hands work
// to a skill of that plugin, so a SKILL.md or a reference of this plugin must
// name one in bold.
export const checkDependencyUsage = (
  input: DependencyUsageInput,
  fs: FileSystem,
): readonly ValidationIssue[] => {
  const handoffs = pluginHandoffNames(input.packageDirectory, fs);
  const isUnused = (plugin: string): boolean => {
    const dependencyDirectory = join(
      dirname(input.packageDirectory),
      `plugin-${plugin.slice(SKILL_PLUGIN_PREFIX.length)}`,
    );
    return pluginSkillDirectories(dependencyDirectory, fs).every(
      (directory) => !handoffs.has(basename(directory)),
    );
  };
  const unusedReason =
    "is declared but no skill of this plugin hands off to one of its skills in bold";

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
