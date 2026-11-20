import {join, relative, resolve} from "node:path";
import {nodeFileSystem} from "@xonovex/script-moon-common/adapters/file-system-node";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {pluginSkillDirectories} from "@xonovex/script-moon-common/package-layout";
import {scanProsePunctuation} from "@xonovex/script-moon-common/prose-punctuation";
import {checkDependencyUsage} from "./dependency-usage.js";
import {checkInternalLinks} from "./internal-links.js";
import {checkReadmeCatalog} from "./readme-catalog.js";
import {
  issue,
  type ValidationIssue,
  type ValidationReport,
} from "./validation.js";

interface PackageManifest {
  readonly dependencies?: Readonly<Record<string, string>>;
}

interface PluginManifest {
  readonly dependencies?: readonly string[];
}

const readJson = (
  path: string,
  issues: ValidationIssue[],
  fs: FileSystem,
): unknown => {
  try {
    return JSON.parse(fs.readText(path)) as unknown;
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    issues.push(
      issue("manifest.invalid", path, `cannot parse JSON: ${message}`),
    );
    return undefined;
  }
};

// The scan covers the whole package: skills, docs, README, and the manifests
// whose descriptions reach a marketplace listing.
const validateProsePunctuation = (
  packageDir: string,
  repositoryRoot: string,
  fs: FileSystem,
): readonly ValidationIssue[] =>
  scanProsePunctuation(packageDir, fs).map(
    ({excerpt, hint, label, line, path}) =>
      issue(
        "prose.punctuation",
        relative(repositoryRoot, join(packageDir, path)),
        `${label} on line ${String(line)}, ${hint}: ${excerpt}`,
      ),
  );

// A plugin holds skills only, so a commands directory is an issue however
// small it is.
const checkNoCommands = (
  packageDir: string,
  repositoryRoot: string,
  fs: FileSystem,
): readonly ValidationIssue[] => {
  const directory = join(packageDir, "commands");
  return fs.isDirectory(directory)
    ? [
        issue(
          "package.commands-present",
          relative(repositoryRoot, directory),
          "a plugin package holds skills only, so its commands directory must not exist",
        ),
      ]
    : [];
};

export const validatePluginPackage = (
  packageDir: string,
  repositoryRoot = resolve(packageDir, "../../../.."),
  fs: FileSystem = nodeFileSystem,
): ValidationReport => {
  const issues: ValidationIssue[] = [];
  const packageManifest = readJson(
    join(packageDir, "package.json"),
    issues,
    fs,
  ) as PackageManifest | undefined;
  const pluginManifest = readJson(
    join(packageDir, ".claude-plugin", "plugin.json"),
    issues,
    fs,
  ) as PluginManifest | undefined;
  const skillNames = pluginSkillDirectories(packageDir, fs).map((directory) =>
    relative(join(packageDir, "skills"), directory),
  );
  issues.push(
    ...checkNoCommands(packageDir, repositoryRoot, fs),
    ...checkInternalLinks(packageDir, repositoryRoot, fs),
    ...validateProsePunctuation(packageDir, repositoryRoot, fs),
    ...checkReadmeCatalog(packageDir, skillNames, repositoryRoot, fs),
    ...checkDependencyUsage(
      {
        packageDirectory: packageDir,
        repositoryRoot,
        pluginDependencies: pluginManifest?.dependencies ?? [],
        packageDependencies: Object.keys(packageManifest?.dependencies ?? {}),
      },
      fs,
    ),
  );
  return {skills: skillNames.length, issues};
};
