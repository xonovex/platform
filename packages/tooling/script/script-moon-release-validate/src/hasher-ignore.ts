import {parse as parseYaml} from "yaml";
import {z} from "zod";

// hasher.walkStrategy is glob, which walks the file system and does not read
// .gitignore, so hasher.ignorePatterns has to restate the exclusions git makes
// via --exclude-standard. The two lists drift silently: a newly ignored
// directory starts being hashed, and a pattern added here starts excluding a
// declared input. This module reports both directions.

export interface DeclaredInput {
  // source names the file the input is declared in, for the failure message.
  readonly source: string;
  readonly path: string;
}

export interface HasherIgnoreInput {
  readonly ignorePatterns: readonly string[];
  // ignoredPaths are workspace relative, as git reports them, with a trailing
  // slash on directories.
  readonly ignoredPaths: readonly string[];
  // declaredInputs holds literal task inputs only. A declared glob cannot be
  // compared against an ignore glob without resolving both against the tree.
  readonly declaredInputs: readonly DeclaredInput[];
  // exemptPrefixes are ignored paths that must stay hashable because a task
  // declares something beneath them.
  readonly exemptPrefixes: readonly string[];
}

// globToRegExp mirrors the subset of glob syntax moon accepts in
// hasher.ignorePatterns: ** spans directories, * stops at a separator.
const wildcardSource: Record<string, string> = {
  "**/": "(?:.*/)?",
  "**": ".*",
  "*": "[^/]*",
};

const globToRegExp = (pattern: string): RegExp => {
  const escaped = pattern.replaceAll(/[.+^${}()|[\]\\]/gu, String.raw`\$&`);
  // One pass over the wildcards: replacing them in sequence would rewrite the
  // `*` inside an already emitted group such as `(?:.*/)?`.
  const source = escaped.replaceAll(
    /\*\*\/|\*\*|\*/gu,
    (match) => wildcardSource[match] ?? match,
  );
  return new RegExp(`^${source}$`, "u");
};

interface CompiledPattern {
  readonly pattern: string;
  readonly regex: RegExp;
}

const matchesAny = (
  path: string,
  compiled: readonly CompiledPattern[],
): boolean => compiled.some(({regex}) => regex.test(path));

// A directory git reports as `dist/` has to match a pattern written for the
// files beneath it, so compare the pattern against a probe path inside it.
const probePaths = (ignoredPath: string): readonly string[] =>
  ignoredPath.endsWith("/")
    ? [`${ignoredPath}probe`, `${ignoredPath}nested/probe`]
    : [ignoredPath];

export const hasherIgnoreFailures = ({
  ignorePatterns,
  ignoredPaths,
  declaredInputs,
  exemptPrefixes,
}: HasherIgnoreInput): readonly string[] => {
  const failures: string[] = [];
  const compiled: readonly CompiledPattern[] = ignorePatterns.map(
    (pattern) => ({
      pattern,
      regex: globToRegExp(pattern),
    }),
  );

  for (const ignoredPath of ignoredPaths) {
    if (exemptPrefixes.some((prefix) => ignoredPath.startsWith(prefix))) {
      continue;
    }
    if (probePaths(ignoredPath).every((probe) => matchesAny(probe, compiled))) {
      continue;
    }
    failures.push(
      `${ignoredPath} is ignored by git but no hasher.ignorePatterns entry covers it: the glob walk will hash it whenever a task input glob reaches it`,
    );
  }

  for (const {source, path} of declaredInputs) {
    const matched = compiled
      .filter(({regex}) => regex.test(path))
      .map(({pattern}) => pattern);
    if (matched.length > 0) {
      failures.push(
        `${source} declares input ${path}, which hasher.ignorePatterns excludes via ${matched.join(", ")}: an ignore pattern drops declared inputs as well as walked ones, so the task stops re-running when that file changes`,
      );
    }
  }

  return failures;
};

const WorkspaceSchema = z.object({
  hasher: z
    .object({
      walkStrategy: z.string().optional(),
      ignorePatterns: z.array(z.string()).optional(),
    })
    .optional(),
});

export interface WorkspaceHasher {
  readonly walkStrategy: string | undefined;
  readonly ignorePatterns: readonly string[];
}

export const parseWorkspaceHasher = (
  text: string,
): WorkspaceHasher | undefined => {
  let input: unknown;
  try {
    input = parseYaml(text);
  } catch {
    return undefined;
  }
  const result = WorkspaceSchema.safeParse(input);
  if (!result.success) return undefined;
  return {
    walkStrategy: result.data.hasher?.walkStrategy,
    ignorePatterns: result.data.hasher?.ignorePatterns ?? [],
  };
};

export interface HasherIgnoreSources {
  // workspaceText is undefined when .moon/workspace.yml is absent. Its absence
  // is not this check's concern: moon cannot run a task without it.
  readonly workspaceText: string | undefined;
  // ignoredPaths is undefined when git could not answer.
  readonly ignoredPaths: readonly string[] | undefined;
  readonly projectFiles: readonly ProjectTaskFile[];
}

// workspaceHasherFailures decides whether the check applies and, when it does,
// reports both directions of drift between .gitignore and ignorePatterns.
export const workspaceHasherFailures = ({
  workspaceText,
  ignoredPaths,
  projectFiles,
}: HasherIgnoreSources): readonly string[] => {
  if (workspaceText === undefined) return [];
  const hasher = parseWorkspaceHasher(workspaceText);
  if (hasher === undefined) {
    return [".moon/workspace.yml has an unreadable hasher section"];
  }
  // The vcs walk applies .gitignore through git, so this alignment only
  // matters while the glob walk is selected.
  if (hasher.walkStrategy !== "glob") return [];
  if (ignoredPaths === undefined) {
    return ["git could not list ignored paths for the hasher check"];
  }

  // A dist directory stays hashable: the bin-permissions tasks declare
  // dist/src/index.js as an input, and an ignore pattern drops declared inputs
  // as well as walked ones, so dist cannot be excluded without breaking them.
  return hasherIgnoreFailures({
    ignorePatterns: hasher.ignorePatterns,
    ignoredPaths,
    declaredInputs: declaredLiteralInputs(projectFiles),
    exemptPrefixes: ignoredPaths.filter((path) => path.endsWith("dist/")),
  });
};

const ProjectSchema = z.object({
  tasks: z
    .record(z.string(), z.object({inputs: z.array(z.string()).optional()}))
    .optional(),
});

export interface ProjectTaskFile {
  // path is the moon.yml path, workspace relative.
  readonly path: string;
  readonly text: string;
}

const isLiteral = (input: string): boolean => !/[*?[\]{}]/u.test(input);

// declaredLiteralInputs resolves the inputs a project declares as plain paths
// to workspace relative form. A glob is skipped: comparing it against an ignore
// glob would need both resolved against the tree.
export const declaredLiteralInputs = (
  files: readonly ProjectTaskFile[],
): readonly DeclaredInput[] => {
  const declared: DeclaredInput[] = [];
  for (const file of files) {
    let input: unknown;
    try {
      input = parseYaml(file.text);
    } catch {
      continue;
    }
    const result = ProjectSchema.safeParse(input);
    if (!result.success) continue;
    const projectDirectory = file.path.slice(0, file.path.lastIndexOf("/"));
    for (const task of Object.values(result.data.tasks ?? {})) {
      for (const entry of task.inputs ?? []) {
        if (!isLiteral(entry)) continue;
        if (entry.startsWith("$")) continue;
        const path = entry.startsWith("/")
          ? entry.slice(1)
          : `${projectDirectory}/${entry}`;
        declared.push({source: file.path, path});
      }
    }
  }
  return declared;
};
