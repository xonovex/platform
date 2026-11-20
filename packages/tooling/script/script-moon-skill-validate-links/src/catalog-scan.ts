import {basename, join} from "node:path";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import {pluginSkillDirectories} from "@xonovex/script-moon-common/package-layout";
import type {FileRead} from "./file-read.js";

export interface MarkdownDocument {
  readonly path: string;
  readonly text: string;
}

export interface PluginSkill {
  readonly name: string;
  readonly skillMarkdown: MarkdownDocument;
  readonly references: readonly MarkdownDocument[];
}

/** One directory under the plugin root, as the link checks need it. */
export interface PluginPackage {
  readonly directory: string;
  readonly name: string;
  readonly hasSkillsDirectory: boolean;
  /** Undefined when the manifest file does not exist. */
  readonly claudeManifest: FileRead | undefined;
  readonly codexManifest: FileRead | undefined;
  readonly packageJson: FileRead;
  readonly skills: readonly PluginSkill[];
  /** The README and docs/ Markdown published beside the skills. */
  readonly documents: readonly MarkdownDocument[];
}

export interface CatalogSnapshot {
  readonly plugins: readonly PluginPackage[];
}

export const pluginRoot = (repoRoot: string): string =>
  join(repoRoot, "packages", "runtime", "plugin");

const readFile = (path: string, fs: FileSystem): FileRead => {
  try {
    return {kind: "read", text: fs.readText(path)};
  } catch (error) {
    return {
      kind: "unreadable",
      message: error instanceof Error ? error.message : String(error),
    };
  }
};

const readDocument = (path: string, fs: FileSystem): MarkdownDocument => ({
  path,
  text: fs.readText(path),
});

const markdownEntries = (directory: string, fs: FileSystem): string[] =>
  fs.isDirectory(directory)
    ? fs
        .readDirectory(directory)
        .filter((entry) => entry.endsWith(".md"))
        .filter((entry) => fs.isFile(join(directory, entry)))
        .toSorted()
        .map((entry) => join(directory, entry))
    : [];

const optionalFile = (path: string, fs: FileSystem): FileRead | undefined =>
  fs.isFile(path) ? readFile(path, fs) : undefined;

const scanPlugin = (directory: string, fs: FileSystem): PluginPackage => ({
  directory,
  name: basename(directory),
  hasSkillsDirectory: fs.isDirectory(join(directory, "skills")),
  claudeManifest: optionalFile(
    join(directory, ".claude-plugin", "plugin.json"),
    fs,
  ),
  codexManifest: optionalFile(
    join(directory, ".codex-plugin", "plugin.json"),
    fs,
  ),
  packageJson: readFile(join(directory, "package.json"), fs),
  skills: pluginSkillDirectories(directory, fs).map((skillDirectory) => ({
    name: basename(skillDirectory),
    skillMarkdown: readDocument(join(skillDirectory, "SKILL.md"), fs),
    references: markdownEntries(join(skillDirectory, "references"), fs).map(
      (path) => readDocument(path, fs),
    ),
  })),
  documents: [
    join(directory, "README.md"),
    ...markdownEntries(join(directory, "docs"), fs),
  ]
    .filter((path) => fs.isFile(path))
    .map((path) => readDocument(path, fs)),
});

/** Reads every package directory under the plugin root, in name order. */
export const scanCatalog = (
  repoRoot: string,
  fs: FileSystem,
): CatalogSnapshot => {
  const root = pluginRoot(repoRoot);
  if (!fs.isDirectory(root)) return {plugins: []};
  return {
    plugins: fs
      .readDirectory(root)
      .toSorted()
      .map((entry) => join(root, entry))
      .filter((directory) => fs.isDirectory(directory))
      .map((directory) => scanPlugin(directory, fs)),
  };
};

/** The Markdown the link checks read: every skill file, then every plugin document. */
export const catalogDocuments = (
  snapshot: CatalogSnapshot,
): readonly MarkdownDocument[] => [
  ...snapshot.plugins.flatMap(({skills}) =>
    skills.flatMap(({references, skillMarkdown}) => [
      skillMarkdown,
      ...references,
    ]),
  ),
  ...snapshot.plugins.flatMap(({documents}) => documents),
];
