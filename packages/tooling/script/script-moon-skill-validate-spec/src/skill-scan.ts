import {dirname, join} from "node:path";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import {
  scanProsePunctuation,
  type ProsePunctuationFinding,
} from "@xonovex/script-moon-common/prose-punctuation";
import {findWorkspaceRoot} from "@xonovex/script-moon-common/workspace";
import {
  checkSkillDrift,
  type DriftLintReport,
} from "@xonovex/script-moon-skill-catalog-common/drift-lints";
import type {ReferenceDocument} from "@xonovex/script-moon-skill-catalog-common/reference-file-links";
import type {CatalogFile} from "./rules/catalog-json.js";
import type {ImplementationFiles} from "./rules/implementation-files.js";

/** What the rules read from a skill directory and the repository around it. */
export interface SkillSnapshot {
  /** Every Markdown file below references/, by "/"-separated path inside it, sorted. */
  readonly referenceDocuments: readonly ReferenceDocument[];
  readonly outputEvals: CatalogFile | undefined;
  readonly triggerEvals: CatalogFile | undefined;
  readonly sources: string | undefined;
  /** The SKILL.md of the directory, which a file target need not be. */
  readonly skillMarkdown: string | undefined;
  readonly implementation: ImplementationFiles;
  readonly punctuation: readonly ProsePunctuationFinding[];
  /** Undefined outside a workspace, which has no budget manifest to measure against. */
  readonly drift: DriftLintReport | undefined;
}

const markdownBelow = (
  directory: string,
  prefix: string,
  fs: FileSystem,
): readonly ReferenceDocument[] =>
  fs.readDirectory(directory).flatMap((name) => {
    const path = join(directory, name);
    const relative = prefix === "" ? name : `${prefix}/${name}`;
    if (fs.isDirectory(path)) return markdownBelow(path, relative, fs);
    return fs.isFile(path) && name.endsWith(".md")
      ? [{name: relative, text: fs.readText(path)}]
      : [];
  });

const readCatalogFile = (
  path: string,
  fs: FileSystem,
): CatalogFile | undefined => {
  if (!fs.isFile(path)) return undefined;
  try {
    return {kind: "read", text: fs.readText(path)};
  } catch (error) {
    return {
      kind: "unreadable",
      message: error instanceof Error ? error.message : String(error),
    };
  }
};

const readOptional = (path: string, fs: FileSystem): string | undefined =>
  fs.isFile(path) ? fs.readText(path) : undefined;

const isSkipped = (name: string): boolean =>
  name === "__pycache__" || name.startsWith(".");

const holdsFile = (directory: string, fs: FileSystem): boolean =>
  fs.isDirectory(directory) &&
  fs.readDirectory(directory).some((name) => {
    if (isSkipped(name)) return false;
    const path = join(directory, name);
    return fs.isDirectory(path) ? holdsFile(path, fs) : fs.isFile(path);
  });

const implementationFilesBelow = (
  directory: string,
  prefix: string,
  fs: FileSystem,
): readonly string[] =>
  fs.readDirectory(directory).flatMap((name) => {
    if (isSkipped(name)) return [];
    const path = join(directory, name);
    const label = prefix === "" ? name : `${prefix}/${name}`;
    if (fs.isDirectory(path)) return implementationFilesBelow(path, label, fs);
    return fs.isFile(path) && /\.(?:mjs|ts)$/u.test(name) ? [label] : [];
  });

const workspaceRootOrUndefined = (
  start: string,
  fs: FileSystem,
): string | undefined => {
  try {
    return findWorkspaceRoot(start, fs);
  } catch {
    return undefined;
  }
};

export const scanSkill = (skillDir: string, fs: FileSystem): SkillSnapshot => {
  const referencesDirectory = join(skillDir, "references");
  // The punctuation scan covers the whole package, not just SKILL.md:
  // references, SOURCES.md, eval fixtures, bundled scripts, asset templates and
  // the manifests all ship prose a reader sees.
  const packageRoot = fs.isFile(join(dirname(skillDir), "package.json"))
    ? dirname(skillDir)
    : skillDir;
  const repositoryRoot = workspaceRootOrUndefined(skillDir, fs);
  return {
    referenceDocuments: fs.isDirectory(referencesDirectory)
      ? markdownBelow(referencesDirectory, "", fs).toSorted((left, right) =>
          left.name < right.name ? -1 : Number(left.name > right.name),
        )
      : [],
    outputEvals: readCatalogFile(join(skillDir, "evals.json"), fs),
    triggerEvals: readCatalogFile(join(skillDir, "eval-queries.json"), fs),
    sources: readOptional(join(skillDir, "SOURCES.md"), fs),
    skillMarkdown: readOptional(join(skillDir, "SKILL.md"), fs),
    implementation: {
      hasScripts: holdsFile(join(skillDir, "scripts"), fs),
      forbidden: implementationFilesBelow(skillDir, "", fs),
    },
    punctuation: scanProsePunctuation(packageRoot, fs),
    drift:
      repositoryRoot === undefined
        ? undefined
        : checkSkillDrift(skillDir, repositoryRoot, fs),
  };
};
