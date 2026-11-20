import {join, resolve} from "node:path";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import type {GitRepository} from "@xonovex/script-moon-common/git";
import {
  packageLayout,
  pluginSkillDirectories,
} from "@xonovex/script-moon-common/package-layout";
import type {InstructionDirectory} from "./instruction-docs.js";
import type {
  FileRead,
  PluginPackageSnapshot,
  PluginPackagesRead,
  ProjectSnapshot,
  ReleaseSnapshot,
} from "./release-snapshot.js";
import type {TagTaskFile} from "./task-inheritance.js";
import {markdownLinkTargets} from "./validate.js";

// The files a project names a config package in, which is how eslint, prettier,
// vitest, and tsc are configured without a dependsOn edge that would cycle.
const CONFIG_FILE_NAMES = [
  "eslint.config.ts",
  "eslint.config.js",
  "prettier.config.ts",
  "prettier.config.js",
  "vitest.config.ts",
  "vite.config.ts",
  "tsconfig.json",
  "tsconfig.build.json",
  "tsconfig.test.json",
] as const;

const errorMessage = (error: unknown): string =>
  error instanceof Error ? error.message : String(error);

const exists = (fs: FileSystem, path: string): boolean =>
  fs.isFile(path) || fs.isDirectory(path);

const readFile = (fs: FileSystem, path: string): FileRead => {
  try {
    return {kind: "text", text: fs.readText(path)};
  } catch (error: unknown) {
    return {kind: "unreadable", reason: errorMessage(error)};
  }
};

// A path that exists but cannot be read, such as a directory, is reported as
// unreadable rather than as absent.
const readIfExists = (fs: FileSystem, path: string): FileRead | undefined =>
  exists(fs, path) ? readFile(fs, path) : undefined;

const textIfFile = (fs: FileSystem, path: string): string | undefined =>
  fs.isFile(path) ? fs.readText(path) : undefined;

// Git cannot answer in a checkout without a git directory. Undefined lets the
// validation report that once rather than one failure per path it would check.
const gitAnswer = <Output>(read: () => Output): Output | undefined => {
  try {
    return read();
  } catch {
    return undefined;
  }
};

const readPluginPackage = (
  fs: FileSystem,
  repositoryRoot: string,
  directory: string,
): PluginPackageSnapshot => ({
  path: directory.slice(repositoryRoot.length + 1),
  packageJson: readIfExists(fs, join(directory, "package.json")),
  neutralManifest: readIfExists(fs, join(directory, "plugin.json")),
  claudeManifest: readIfExists(
    fs,
    join(directory, ".claude-plugin/plugin.json"),
  ),
  codexManifest: readIfExists(fs, join(directory, ".codex-plugin/plugin.json")),
  skillCount: pluginSkillDirectories(directory, fs).length,
  skillsPathExists: exists(fs, join(directory, "skills")),
});

const readPluginPackages = (
  fs: FileSystem,
  repositoryRoot: string,
): PluginPackagesRead => {
  const root = join(repositoryRoot, "packages/runtime/plugin");
  try {
    const directories = fs
      .readDirectory(root)
      .map((name) => join(root, name))
      .filter((path) => fs.isDirectory(path))
      .toSorted();
    return {
      kind: "listed",
      packages: directories.map((directory) =>
        readPluginPackage(fs, repositoryRoot, directory),
      ),
    };
  } catch (error: unknown) {
    return {kind: "unlisted", reason: errorMessage(error)};
  }
};

const readTagTaskFiles = (
  fs: FileSystem,
  repositoryRoot: string,
): readonly TagTaskFile[] => {
  const directory = join(repositoryRoot, ".moon/tasks");
  if (!fs.isDirectory(directory)) return [];
  return fs
    .readDirectory(directory)
    .toSorted()
    .filter((entry) => entry.endsWith(".yml"))
    .map((entry) => ({
      name: entry,
      text: fs.readText(join(directory, entry)),
    }));
};

const readProject = (
  fs: FileSystem,
  repositoryRoot: string,
  directory: string,
): ProjectSnapshot | undefined => {
  const moonPath = join(directory, "moon.yml");
  if (!fs.isFile(moonPath)) return undefined;
  return {
    path: moonPath.slice(repositoryRoot.length + 1),
    text: fs.readText(moonPath),
    packageJsonText: textIfFile(fs, join(directory, "package.json")),
    configTexts: CONFIG_FILE_NAMES.map((name) =>
      textIfFile(fs, join(directory, name)),
    ).filter((text): text is string => text !== undefined),
  };
};

const readInstructionDirectory = (
  fs: FileSystem,
  repositoryRoot: string,
  directory: string,
): InstructionDirectory => ({
  path: directory.slice(repositoryRoot.length + 1) || ".",
  hasAgentsFile: fs.isFile(join(directory, "AGENTS.md")),
  claudeText: textIfFile(fs, join(directory, "CLAUDE.md")),
});

/** Reads every input of the release validation from the repository. */
export const readReleaseSnapshot = (
  repositoryRoot: string,
  fs: FileSystem,
  git: Pick<GitRepository, "trackedFiles" | "ignoredPaths">,
): ReleaseSnapshot => {
  const layout = packageLayout(repositoryRoot, fs);
  const projectDirectories = layout.packages.flatMap((directory) => [
    directory,
    ...pluginSkillDirectories(directory, fs),
  ]);
  const readme = readFile(fs, join(repositoryRoot, "README.md"));
  const trackedPaths = gitAnswer(() => git.trackedFiles());

  return {
    claudeMarketplace: readFile(
      fs,
      join(repositoryRoot, ".claude-plugin/marketplace.json"),
    ),
    codexMarketplace: readFile(
      fs,
      join(repositoryRoot, ".agents/plugins/marketplace.json"),
    ),
    pluginPackages: readPluginPackages(fs, repositoryRoot),
    lockfile: readFile(fs, join(repositoryRoot, "package-lock.json")),
    readme,
    existingLinkTargets: new Set(
      readme.kind === "text"
        ? markdownLinkTargets(readme.text).filter((target) =>
            exists(fs, resolve(repositoryRoot, target)),
          )
        : [],
    ),
    releaseWorkflow: readFile(
      fs,
      join(repositoryRoot, ".github/workflows/release.yml"),
    ),
    tagTaskFiles: readTagTaskFiles(fs, repositoryRoot),
    projects: projectDirectories
      .map((directory) => readProject(fs, repositoryRoot, directory))
      .filter((project) => project !== undefined),
    workspaceText: textIfFile(fs, join(repositoryRoot, ".moon/workspace.yml")),
    ignoredPaths: gitAnswer(() => git.ignoredPaths()),
    trackedPaths:
      trackedPaths === undefined ? undefined : new Set(trackedPaths),
    packageManifests: layout.packages
      .filter((directory) => fs.isFile(join(directory, "package.json")))
      .map((directory) => ({
        directory: directory.slice(repositoryRoot.length + 1),
        manifest: readFile(fs, join(directory, "package.json")),
      })),
    instructionDirectories: [
      repositoryRoot,
      join(repositoryRoot, ".moon"),
      ...layout.groups,
      ...projectDirectories,
    ].map((directory) =>
      readInstructionDirectory(fs, repositoryRoot, directory),
    ),
    requiredInstructionDirectories: layout.groups.map((group) =>
      group.slice(repositoryRoot.length + 1),
    ),
  };
};
