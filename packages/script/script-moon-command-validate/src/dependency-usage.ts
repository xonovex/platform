import {join, relative} from "node:path";
import {
  nodeFileSystem,
  type FileSystem,
} from "@xonovex/script-moon-common/file-system";
import {issue, type ValidationIssue} from "./validation.js";

const SKILL_PLUGIN_PREFIX = "xonovex-skill-";
const SKILL_PACKAGE_PREFIX = "@xonovex/skill-";

const isBoundary = (character: string | undefined): boolean =>
  character === undefined || !/[\p{Letter}\p{Number}-]/u.test(character);

// The subject is matched as a whole word so an install line naming the plugin
// id cannot vouch for the dependency it installs.
const mentionsSubject = (source: string, subject: string): boolean => {
  const haystack = source.toLowerCase();
  for (
    let index = haystack.indexOf(subject);
    index !== -1;
    index = haystack.indexOf(subject, index + 1)
  ) {
    if (
      isBoundary(haystack[index - 1]) &&
      isBoundary(haystack[index + subject.length])
    ) {
      return true;
    }
  }
  return false;
};

const documentationSources = (
  packageDirectory: string,
  commandNames: readonly string[],
  fs: FileSystem,
): readonly string[] => {
  const readme = join(packageDirectory, "README.md");
  const commandDirectory = join(packageDirectory, "commands");
  return [
    ...(fs.isFile(readme) ? [fs.readText(readme)] : []),
    ...commandNames
      .map((name) => join(commandDirectory, `${name}.md`))
      .filter((path) => fs.isFile(path))
      .map((path) => fs.readText(path)),
  ];
};

export interface DependencyUsageInput {
  readonly packageDirectory: string;
  readonly repositoryRoot: string;
  readonly commandNames: readonly string[];
  readonly pluginDependencies: readonly string[];
  readonly packageDependencies: readonly string[];
  readonly delegatedPlugins: ReadonlySet<string>;
}

// checkDependencyUsage is the reverse of the delegation checks: those require a
// dependency for every delegation, this requires a reason for every dependency.
// A skill reached through another skill's routing carries no Delegation block,
// so naming it in the README or a command document counts as that reason.
export const checkDependencyUsage = (
  input: DependencyUsageInput,
  fs: FileSystem = nodeFileSystem,
): readonly ValidationIssue[] => {
  const sources = documentationSources(
    input.packageDirectory,
    input.commandNames,
    fs,
  );
  const isUnused = (plugin: string): boolean => {
    if (input.delegatedPlugins.has(plugin)) return false;
    const subject = plugin.slice(SKILL_PLUGIN_PREFIX.length);
    return sources.every((source) => !mentionsSubject(source, subject));
  };
  const unusedReason =
    "is declared but no command delegates to it and no document names it";

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
